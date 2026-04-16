package ntwk

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	Url               *URL
	headers           *map[string]string
	max_num_redirects *int
}

func GetRequest(urlStr string, headers *map[string]string, max_num_redirects *int) (*Request, error) {
	req := &Request{}
	err := req.Parse(urlStr, headers, max_num_redirects)
	return req, err
}

func (req *Request) Parse(urlStr string, headers *map[string]string, max_num_redirects *int) error {
	req.Url = &URL{}
	if err := req.Url.Parse(urlStr); err != nil {
		return err
	}

	if headers == nil {
		headers = &map[string]string{}
	}
	(*headers)["Host"] = req.Url.Host
	(*headers)["User-Agent"] = "browsy"
	(*headers)["Connection"] = "close"

	req.headers = headers
	req.max_num_redirects = max_num_redirects

	return nil
}

func (req *Request) Do1() (string, error) {
	res, err := req.Do()
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var reader io.Reader
	if res.Reader != nil {
		reader = res.Reader
	} else {
		reader = res.Body
	}

	if res.Headers == nil {
		return readContent(reader)
	}

	lengthStr, exists := (*res.Headers)["content-length"]
	if !exists {
		return readContent(reader)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("error: converting content-length to int: %s", err)
	}

	return readWithContentLength(reader, length)
}

func (req *Request) Do() (*Response, error) {
	res := Response{}

	if req.Url.Scheme == DATA {
		return &res, req.dataRequest(&res)
	}

	if req.Url.Scheme == FILE {
		return &res, req.fileRequest(&res)
	}

	responseHeaders := make(map[string]string)
	res.Headers = &responseHeaders

	for {
		conn, err := net.Dial("tcp", net.JoinHostPort(req.Url.Host, strconv.Itoa(req.Url.Port)))

		if err != nil {
			return nil, fmt.Errorf("eeror dialing a connection: %s", err)
		}

		if req.Url.Scheme == HTTPS {
			conn, err = getSecureConnection(conn, req.Url.Host)
			if err != nil {
				return nil, err
			}
		}

		res.Body = conn

		request := req.constructHTTPRequest()

		fmt.Println("------------Request------------")
		fmt.Println(request)

		if _, err = conn.Write([]byte(request)); err != nil {
			return nil, fmt.Errorf("Erorr writing bytes: %s", err)
		}

		responseReader := bufio.NewReader(conn)
		res.Reader = responseReader

		statusLine, err := responseReader.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("Error reading the status line: %s", err)
		}

		statusInfo := strings.SplitN(string(statusLine), " ", 3)
		if len(statusInfo) < 3 {
			return nil, fmt.Errorf("Invalid status info")
		}

		fmt.Printf("Version: %s, Status: %s, explanation: %s\n", statusInfo[0], statusInfo[1], statusInfo[2])

		for {
			rawData, err := responseReader.ReadBytes('\n')
			line := string(rawData)
			if err != nil {
				if err != io.EOF {
					return nil, fmt.Errorf("Error reading response: %s", err)
				}
				if len(rawData) > 0 {
					fmt.Printf("%s\n", line)
				}
				break
			}

			if line == "\r\n" {
				break
			}

			header, value, found := strings.Cut(line, ":")
			if found {
				header = strings.ToLower(header)
				value = strings.TrimSpace(value)
				responseHeaders[header] = value
				fmt.Printf("%s:%s\n", header, responseHeaders[header])
				continue
			}

			fmt.Printf("%s\n", line)
		}

		if _, exists := responseHeaders["transfer-encoding"]; exists {
			return nil, fmt.Errorf("'transfer-encoding' is Present")
		}

		if _, exists := responseHeaders["content-encoding"]; exists {
			return nil, fmt.Errorf("'content-encoding' is Present")
		}

		if statusInfo[1][0] != '3' {
			break
		}

		if *req.max_num_redirects == 0 {
			log.Println("error: reached max number of redirection")
			break
		}

		log.Println("Redirecting...")

		location, exists := responseHeaders["location"]
		if !exists || location == "" {
			log.Println("error in redirect: location header is empty in the")
			break
		}

		if location[0] == '/' {
			req.Url.Path = location
		} else {
			if err = req.Url.Parse(location); err != nil {
				log.Println("error: parsing the new locaion:", location)
				break
			}
		}

		conn.Close()
		*req.max_num_redirects -= 1
	}

	return &res, nil

}

func (req *Request) constructHTTPRequest() string {
	var request strings.Builder
	fmt.Fprintf(&request, "GET %s HTTP/1.1\r\n", req.Url.Path)

	if req.headers != nil {
		for header, value := range *req.headers {
			fmt.Fprintf(&request, "%s: %s\r\n", header, value)
		}
	}

	request.WriteString("\r\n")
	return request.String()
}

func getSecureConnection(conn net.Conn, host string) (*tls.Conn, error) {
	secureConn := tls.Client(conn, &tls.Config{
		ServerName: host,
	})

	if err := secureConn.Handshake(); err != nil {
		return nil, fmt.Errorf("Error in the handshake: %s", err)
	}

	return secureConn, nil
}

func readWithContentLength(reader io.Reader, contentLength int) (string, error) {
	content := make([]byte, contentLength)
	n, err := io.ReadFull(reader, content)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return formatOutput(string(content[:n])), nil
		}
		return "", fmt.Errorf("Error reading with content-length: %s", err)
	}

	return formatOutput(string(content[:n])), nil
}

func readContent(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("Error reading the content: %s", err)
	}

	return formatOutput(string(content)), nil
}

func (req *Request) fileRequest(res *Response) error {
	info, err := os.Stat(req.Url.Path)
	if err != nil {
		return fmt.Errorf("Error parsing the path: %s", err)
	}
	if info.IsDir() {
		contents, err := os.ReadDir(req.Url.Path)
		res.Body = io.NopCloser(strings.NewReader(fmt.Sprint(contents)))
		return err

	}

	res.Body = io.NopCloser(strings.NewReader(fmt.Sprint(info)))

	return nil
}

func (req *Request) dataRequest(res *Response) error {
	content := req.Url.Path
	metadata, data, found := strings.Cut(content, ",")
	if !found {
		return fmt.Errorf("no data provided")
	}

	if metadata != "" {

	}

	res.Body = io.NopCloser(strings.NewReader(data))
	return nil
}

var lt *regexp.Regexp = regexp.MustCompile("&lt;")
var gt *regexp.Regexp = regexp.MustCompile("&gt;")

func formatOutput(content string) string {
	return gt.ReplaceAllString(lt.ReplaceAllString(content, "<"), ">")
}
