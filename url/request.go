package url

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
)

func (url *URL) Request(headers *map[string]string) (string, error) {

	if url.Scheme == DATA {
		return url.dataRequest()
	}

	if url.Scheme == FILE {
		return url.fileRequest()
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(url.Host, strconv.Itoa(url.Port)))

	if err != nil {
		return "", fmt.Errorf("Error dialing a connection: %s", err)
	}

	defer conn.Close()

	if url.Scheme == HTTPS {
		secureConn := tls.Client(conn, &tls.Config{
			ServerName: url.Host,
		})

		if err = secureConn.Handshake(); err != nil {
			return "", fmt.Errorf("Error in the handshake: %s", err)
		}

		conn = secureConn
	}

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", url.Path)
	request += fmt.Sprintf("Host: %s\r\n", url.Host)
	request += "Connection: close\r\n"
	request += "User-Agent: browsy\r\n"

	if headers != nil {
		for header, value := range *headers {
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

func (url *URL) fileRequest() (string, error) {
	info, err := os.Stat(url.Path)
	if err != nil {
		return "", fmt.Errorf("Error parsing the path: %s", err)
	}
	if info.IsDir() {
		contents, err := os.ReadDir(url.Path)
		return fmt.Sprintf("%v", contents), err
	}

	return fmt.Sprintf("%v", info), nil
}

func (url *URL) dataRequest() (string, error) {
	content := url.Path
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
