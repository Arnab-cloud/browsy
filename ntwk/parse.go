package ntwk

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

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
func (url *URL) Parse(urlStr string) error {
	urlStr = strings.Trim(urlStr, "\"")

	if err := url.parseScheme(urlStr); err != nil {
		return err
	}

	if err := url.validatePathAndScheme(); err != nil {
		return err
	}

	if url.Scheme == DATA || url.Scheme == FILE {
		return nil
	}

	return url.parseHTTPPath()
}

func (url *URL) parseHTTPPath() error {
	urlStr, portStr, portFound := strings.Cut(url.Path, ":")

	if !portFound {
		url.Port = url.Scheme.GetDefaultPort()
		url.Host, url.Path, _ = strings.Cut(urlStr, "/")
		url.Path = "/" + url.Path

		return nil
	}

	portStr, path, _ := strings.Cut(portStr, "/")
	parsedPort, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("Error parsing the port: %s", err)
	}

	url.Port = parsedPort
	url.Host = urlStr
	url.Path = "/" + path

	return nil
}

// The url.Path is parsed, extra
func (url *URL) validatePathAndScheme() error {
	switch url.Scheme {
	case HTTP, HTTPS, FILE:
		cleanedPath, found := strings.CutPrefix(url.Path, "//")
		if !found {
			return fmt.Errorf("invalid path: %s", url.Path)
		}
		cleanedPath = strings.TrimLeft(cleanedPath, "/")
		if url.Scheme == FILE {
			cleanedPath = filepath.Clean(cleanedPath)
		}
		url.Path = cleanedPath
		return nil
	case DATA:
		return nil
	default:
		return fmt.Errorf("scheme not supported")
	}
}

func (url *URL) parseScheme(urlStr string) error {
	schemePart, path, schemeFound := strings.Cut(urlStr, ":")
	if !schemeFound {
		return fmt.Errorf("scheme not found")
	}
	url.Scheme = SchemeType(schemePart)
	url.Path = path
	return nil
}
