package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Arnab-cloud/browsy/request"
)

func runURL(t *testing.T, input string) {
	req, err := request.GetRequest(input, nil, nil)

	if err != nil {
		t.Fatalf("parse failed for %s: %v", input, err)
	}

	content, err := req.Do()
	if err != nil {
		t.Fatalf("request failed for %s: %v", input, err)
	}

	if len(strings.TrimSpace(content)) == 0 {
		t.Errorf("empty content for %s", input)
	}

	t.Log(content)
}

func TestHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HTTP test in short mode")
	}

	runURL(t, "http://example.org")
}

func TestHTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HTTPS test in short mode")
	}

	runURL(t, "https://example.org")
}

func TestHTTP_LocalFileServer(t *testing.T) {
	// serve files from testdata/
	fs := http.FileServer(http.Dir("testdata"))

	server := httptest.NewServer(fs)
	defer server.Close()

	req, err := request.GetRequest(server.URL, nil, nil)

	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	content, err := req.Do()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !strings.Contains(content, "<html>") {
		t.Errorf("invalid HTML response")
	}

	t.Log(content)
}

func TestFile(t *testing.T) {
	// create temp file
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "hello from file"
	tmpFile.WriteString(content)
	tmpFile.Close()

	runURL(t, "file://"+tmpFile.Name())
	runURL(t, "file:///"+tmpFile.Name())
	runURL(t, "file:////"+tmpFile.Name())
}

func TestData(t *testing.T) {
	runURL(t, "data:text/plain,HelloWorld")
}
