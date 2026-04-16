package ntwk

import (
	"bufio"
	"io"
)

type Response struct {
	// Req     *Request
	Headers *map[string]string
	Body    io.ReadCloser
	Reader  *bufio.Reader
}
