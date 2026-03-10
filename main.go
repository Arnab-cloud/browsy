package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
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

	if !strings.Contains(urlStr, "/") {
		urlStr += "/"
	}

	url.Scheme = SchemeType(urlParts[0])
	urlStr = urlParts[1]

	urlParts = strings.Split(urlStr, "/")
	url.Host = urlParts[0]
	url.Path = strings.Join(urlParts[1:], "/") + "/"
	url.Port = 80
}

func (url *URL) Request() (string, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(url.Host, fmt.Sprintf("%d", url.Port)))

	if err != nil {
		return "", fmt.Errorf("Error dialing a connection: %s", err)
	}

	defer conn.Close()

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

	statusInfo := strings.Split(string(statusLine), " ")
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
		header, value, found := strings.Cut(line, ": ")
		if found {
			responseHeaders[header] = strings.Trim(value, "\r\n")
			fmt.Printf("%s:%s\n", header, responseHeaders[header])
			continue
		}

		fmt.Printf("%s\n", line)
	}

	if _, exists := responseHeaders["Transfer-Encoding"]; exists {
		return "", fmt.Errorf("'Transfer-Encoding' is Present")
	}

	if _, exists := responseHeaders["Content-Encoding"]; exists {
		return "", fmt.Errorf("'Content-Encoding' is Present")
	}

	content, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("Error reading the content: %s", err)
	}

	return string(content), nil

}

func main() {
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
