package rest

import (
	"errors"
	"net/http"
)

type BadReader struct {
}

func (br BadReader) Read(p []byte) (int, error) {
	return 0, errors.New("some error")
}

type BadResponseWriter struct {
	http.ResponseWriter
}

func (bw BadResponseWriter) Write(p []byte) (int, error) {
	return 0, errors.New("some error")
}
