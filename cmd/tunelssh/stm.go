package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/balibuild/tunnelssh/cli"
)

// Status todo
type Status int

// STM
const (
	ReadNone Status = iota
	ReadCR
	ReadCRLF
	ReadCRLFCR
	ReadCRLFCRLF
)

// StateMachineCONNECT todo
func StateMachineCONNECT(conn net.Conn) (*http.Response, error) {
	var status Status
	var buf bytes.Buffer
	buf.Grow(1024)
	var b [16]byte
	for {
		i, err := conn.Read(b[0:1])
		if err != nil {
			return nil, err
		}
		if i != 1 {
			return nil, errors.New("connection is closed")
		}
		_ = buf.WriteByte(b[0])
		switch b[0] {
		case '\r':
			switch status {
			case ReadCRLF:
				status = ReadCRLFCR
			default:
				status = ReadCR
			}
		case '\n':
			switch status {
			case ReadCR:
				status = ReadCRLF
			case ReadCRLFCR:
				status = ReadCRLFCRLF
			default:
				status = ReadNone
			}
		default:
			status = ReadNone
		}
		if status == ReadCRLFCRLF {
			break
		}
	}
	resp := &http.Response{}
	br := bufio.NewReader(&buf)
	tp := textproto.NewReader(br)
	// Parse the first line of the response.
	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	i := strings.IndexByte(line, ' ')
	if i == -1 {
		return nil, cli.ErrorCat("malformed HTTP response", line)
	}
	resp.Proto = line[:i]
	resp.Status = strings.TrimLeft(line[i+1:], " ")
	statusCode := resp.Status
	if i := strings.IndexByte(resp.Status, ' '); i != -1 {
		statusCode = resp.Status[:i]
	}
	if len(statusCode) != 3 {
		return nil, cli.ErrorCat("malformed HTTP status code", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return nil, cli.ErrorCat("malformed HTTP status code", statusCode)
	}
	var ok bool
	if resp.ProtoMajor, resp.ProtoMinor, ok = http.ParseHTTPVersion(resp.Proto); !ok {
		return nil, cli.ErrorCat("malformed HTTP version", resp.Proto)
	}
	// Parse the response headers.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	resp.Header = http.Header(mimeHeader)
	return resp, nil
}
