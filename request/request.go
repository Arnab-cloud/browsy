package request

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Arnab-cloud/browsy/url"
)

type Request struct {
	Url               *url.URL
	headers           *map[string]string
	max_num_redirects *int
}

func GetRequest(urlStr string, headers *map[string]string, max_num_redirects *int) (*Request, error) {
	req := &Request{}
	err := req.Parse(urlStr, headers, max_num_redirects)
	return req, err
}

func (req *Request) Parse(urlStr string, headers *map[string]string, max_num_redirects *int) error {
	req.Url = &url.URL{}
	if err := req.Url.Parse(urlStr); err != nil {
		return err
	}

	req.headers = headers
	req.max_num_redirects = max_num_redirects

	return nil
}

func (req *Request) Do() (string, error) {
	return req.Request()
}

func (req *Request) Request() (string, error) {

	if req.Url.Scheme == url.DATA {
		return req.dataRequest()
	}

	if req.Url.Scheme == url.FILE {
		return req.fileRequest()
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(req.Url.Host, strconv.Itoa(req.Url.Port)))

	if err != nil {
		return "", fmt.Errorf("Error dialing a connection: %s", err)
	}

	defer conn.Close()

	if req.Url.Scheme == url.HTTPS {
		secureConn := tls.Client(conn, &tls.Config{
			ServerName: req.Url.Host,
		})

		if err = secureConn.Handshake(); err != nil {
			return "", fmt.Errorf("Error in the handshake: %s", err)
		}

		conn = secureConn
	}

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", req.Url.Path)
	request += fmt.Sprintf("Host: %s\r\n", req.Url.Host)
	request += "Connection: close\r\n"
	request += "User-Agent: browsy\r\n"

	if req.headers != nil {
		for header, value := range *req.headers {
			request += fmt.Sprintf("%s: %s\r\n", header, value)
		}
	}

	request += "\r\n"

	fmt.Println("------------Request------------")
	fmt.Println(request)

	if _, err = conn.Write([]byte(request)); err != nil {
		return "", fmt.Errorf("Erorr writing bytes: %s", err)
	}

	responseHeaders := make(map[string]string)

	reader := bufio.NewReader(conn)

	fmt.Println("------------Response------------")
	statusLine, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("Error reading the status line: %s", err)
	}

	statusInfo := strings.SplitN(string(statusLine), " ", 3)
	if len(statusInfo) < 3 {
		return "", fmt.Errorf("Invalid status info")
	}

	fmt.Printf("Version: %s, Status: %s, explanation: %s\n", statusInfo[0], statusInfo[1], statusInfo[2])

	for {
		rawData, err := reader.ReadBytes('\n')
		line := string(rawData)
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf("Error reading response: %s", err)
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
		return "", fmt.Errorf("'transfer-encoding' is Present")
	}

	if _, exists := responseHeaders["content-encoding"]; exists {
		return "", fmt.Errorf("'content-encoding' is Present")
	}

	if contentLenStr, exists := responseHeaders["content-length"]; exists {
		contenLength, err := strconv.Atoi(contentLenStr)
		if err == nil {
			return readWithContentLength(reader, contenLength)
		}
	}

	return readContent(reader)

}

func readWithContentLength(reader io.Reader, contentLength int) (string, error) {
	content := make([]byte, contentLength)
	_, err := io.ReadFull(reader, content)
	if err != nil {
		return "", fmt.Errorf("Error reading with content-length: %s", err)
	}
	return formatOutput(string(content)), nil
}

func readContent(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("Error reading the content: %s", err)
	}

	return formatOutput(string(content)), nil
}

func (req *Request) fileRequest() (string, error) {
	info, err := os.Stat(req.Url.Path)
	if err != nil {
		return "", fmt.Errorf("Error parsing the path: %s", err)
	}
	if info.IsDir() {
		contents, err := os.ReadDir(req.Url.Path)
		return fmt.Sprintf("%v", contents), err
	}

	return fmt.Sprintf("%v", info), nil
}

func (req *Request) dataRequest() (string, error) {
	content := req.Url.Path
	metadata, data, found := strings.Cut(content, ",")
	if !found {
		return "", fmt.Errorf("no data provided")
	}

	if metadata != "" {

	}

	return data, nil
}

var lt *regexp.Regexp = regexp.MustCompile("&lt;")
var gt *regexp.Regexp = regexp.MustCompile("&gt;")

func formatOutput(content string) string {
	return gt.ReplaceAllString(lt.ReplaceAllString(content, "<"), ">")
}
