package sshproxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

func NewTypeWriterReadCloser(r io.ReadCloser) io.ReadCloser {
	return &TypeWriterReadCloser{ReadCloser: r, time: time.Now()}
}

type TypeWriterReadCloser struct {
	io.ReadCloser

	time   time.Time
	buffer bytes.Buffer
}

func sanitize(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "<br/>", -1)
	s = strings.Replace(s, "'", "\\'", -1)
	s = strings.Replace(s, "\b", "<backspace>", -1)
	return s
}

func (lr *TypeWriterReadCloser) Read(p []byte) (n int, err error) {
	n, err = lr.ReadCloser.Read(p)

	now := time.Now()
	lr.buffer.WriteString(fmt.Sprintf(".wait(%d)", int(now.Sub(lr.time).Seconds()*1000)))
	lr.buffer.WriteString(fmt.Sprintf(".put('%s')", sanitize(string(p[:n]))))
	lr.time = now

	return n, err
}

func (lr *TypeWriterReadCloser) String() string {
	return lr.buffer.String()
}

func (lr *TypeWriterReadCloser) Close() error {
	return lr.ReadCloser.Close()
}

func NewLogReadCloser(r io.ReadCloser) io.ReadCloser {
	return &LogReadCloser{ReadCloser: r}
}

type LogReadCloser struct {
	io.ReadCloser
}

func (lr *LogReadCloser) Close() error {
	return lr.ReadCloser.Close()
}

func (lr *LogReadCloser) Read(p []byte) (n int, err error) {
	n, err = lr.ReadCloser.Read(p)
	log.Print(string(p[:n]))
	return n, err
}
