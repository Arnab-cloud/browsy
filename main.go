package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"crypto/tls"
)

type SchemeType string

const (
	HTTP  SchemeType = "http"
	HTTPS SchemeType = "https"
)

type URL struct {
	Host   string
	Path   string
	Scheme SchemeType
	Port   int
}

func (url *URL) ParseURL(urlStr string) {
	urlParts := strings.Split(urlStr, "://")
	if len(urlParts) < 2 {
		return
	}

	url.Scheme = SchemeType(urlParts[0])

	switch url.Scheme {
	case HTTP:
		url.Port = 80
	case HTTPS:
		url.Port = 443
	default:
		url.Port = 80
	}

	urlStr = urlParts[1]

	urlParts = strings.Split(urlStr, "/")
	url.Host = urlParts[0]
	url.Path = "/" + strings.Join(urlParts[1:], "/")
}

func (url *URL) Request() (string, error) {
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

	request := fmt.Sprintf("GET %s HTTP/1.0\r\n", url.Path)
	request += fmt.Sprintf("Host: %s\r\n", url.Host)
	request += "\r\n"

	if _, err = conn.Write([]byte(request)); err != nil {
		return "", fmt.Errorf("Erorr writing bytes: %s", err)
	}

	responseHeaders := make(map[string]string)

	reader := bufio.NewReader(conn)

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
	return string(content), nil
}

func readContent(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("Error reading the content: %s", err)
	}

	return string(content), nil
}

func sendRequest() {
	if len(os.Args) < 2 {
		log.Fatalf("Provide a URL\n")
	}

	url := URL{}

	url.ParseURL(os.Args[1])
	fmt.Printf("Parsed url: %v\n\n", url)
	fmt.Printf("------------Response------------\n")

	content, err := url.Request()
	if err != nil {
		log.Fatalf("Erorr: %s\n", err)
	}

	fmt.Printf("HTML:\n%s\n", content)
}

func parseHTMLTag(content string) {
	inTag := false
	for _, ch := range content {
		switch ch {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				fmt.Printf("%c", ch)
			}
		}
	}
}

func main() {
	sendRequest()

	// content := "<!doctype html><html lang=\"en\">	<head>		<title>Example Domain</title>		<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">		<style>			body {			background:#eee;width:60vw;margin:15vh auto;font-family:system-ui,sans-serif}h1{font-size:1.5em}div{opacity:0.8}a:link,a:visited{color:#348}		</style>	</head>	<body><div><h1>Example Domain</h1><p>This domain is for use in documentation examples without needing permission. Avoid use in operations.</p><p><a href=\"https://iana.org/domains/example\">Learn more</a></p></div>	</body></html>"

	// parseHTMLTag(content)
}
