package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Arnab-cloud/browsy/url"
)

func runURL(t *testing.T, input string) {
	u := url.URL{}

	if err := u.Parse(input); err != nil {
		t.Fatalf("parse failed for %s: %v", input, err)
	}

	content, err := u.Request(nil)
	if err != nil {
		t.Fatalf("request failed for %s: %v", input, err)
	}

	if len(strings.TrimSpace(content)) == 0 {
		t.Errorf("empty content for %s", input)
	}
}

func TestHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HTTP test in short mode")
	}

	runURL(t, "http://example.com")
}

func TestHTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping HTTPS test in short mode")
	}

	runURL(t, "https://example.com")
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
}

func TestData(t *testing.T) {
	runURL(t, "data:text/plain,HelloWorld")
}
