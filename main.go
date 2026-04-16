package main

import (
	"fmt"
	"log"
	"os"

	url "github.com/Arnab-cloud/browsy/url"
)

func sendRequest() {
	if len(os.Args) < 2 {
		log.Fatalf("Provide a URL\n")
	}

	url := url.URL{}

	if err := url.Parse(os.Args[1]); err != nil {
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
