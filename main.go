package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"crypto/tls"
)

type SchemeType string

const (
	HTTP  SchemeType = "http"
	HTTPS SchemeType = "https"
	File  SchemeType = "file"
	data  SchemeType = "data"
)

func (schema SchemeType) GetDefaultPort() int {
	switch schema {
	case HTTP:
		return 80
	case HTTPS:
		return 443
	case File:
		return 0
	case data:
		return 0
	default:
		return 80
	}
}

type URL struct {
	Host   string
	Path   string
	Scheme SchemeType
	Port   int
}

// func extractScheme(url *string) (SchemeType, error) {

// }

// urlStr is the url string to be parsed.
// schemes accepted are HTTP, HTTPS & File
//
// For HTTPS & HTTPS, after successfully parsing the URL.Scheme, URL.Host will be set as ususal.
// URL.Port = 80 for HTTP, and URL.Port = 443 for HTTPS
// IF Custom port is porvided that will be used instead of default port.
// URL.Path will be set to the path after the port (is exists).
//
// For File scheme, URL.Host will be localhost
// URL.Path will be the directory or file path as provided in urlStr
// URL.Port will be 0 and ignored
// URL.Scheme will be File.
// eg.. file:///path/to/file
// here the path will be /path/to/file .
//
// File urls with with hostname are not accecpted.
// eg.. file:///localhost/path/to/file will result unexpected behaviour
func (url *URL) ParseURL(urlStr string) error {

	urlStr = strings.Trim(urlStr, "\"")

	schemeStr, urlStr, schemeFound := strings.Cut(urlStr, "://")
	if !schemeFound {
		fmt.Println("INFO: '://' not found")
		fmt.Println("parsing the file as file path")

		url.SetFilePath(schemeStr)
		return nil
	}

	scheme := SchemeType(schemeStr)
	path := ""
	hostName := ""
	port := 0

	if scheme == File {
		url.SetFilePath(urlStr)
		return nil
	}

	urlStr, portStr, portFound := strings.Cut(urlStr, ":")

	if !portFound {
		port = scheme.GetDefaultPort()
		hostName, path, _ = strings.Cut(urlStr, "/")
	} else {
		portStr, path, _ = strings.Cut(portStr, "/")
		parsedPort, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("Error parsing the port: %s", err)
		}

		port = parsedPort
		hostName = urlStr
	}

	url.Scheme = scheme
	url.Port = port
	url.Host = hostName
	url.Path = "/" + path

	return nil
}

func (url *URL) SetFilePath(path string) {
	url.Scheme = File
	if path[0] == '/' {
		url.Path = filepath.Clean(path[1:])
	} else {
		url.Path = filepath.Clean(path)
	}
	url.Host = "localhost"
	url.Port = 0
}

func (url *URL) FileRequest() (string, error) {
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

func (url *URL) Request(headers *map[string]string) (string, error) {

	if url.Scheme == File {
		return url.FileRequest()
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

	if err := url.ParseURL(os.Args[1]); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Printf("Parsed url: %v\n\n", url)

	content, err := url.Request(nil)
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
