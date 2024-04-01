package main

import (
	"bytes"
	"log"
	"net/http"
)

type responseWriter struct {
	Body         bytes.Buffer
	CustomHeader http.Header
	StatusCode   int
}

func (crw *responseWriter) Header() http.Header {
	return crw.CustomHeader
}

func (crw *responseWriter) Write(b []byte) (int, error) {
	return crw.Body.Write(b)
}

func (crw *responseWriter) WriteHeader(statusCode int) {
	crw.StatusCode = statusCode
}

type fallbackFileServer struct {
	primary   http.Handler
	secondary http.Handler
}

func (s *fallbackFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
	rw := responseWriter{ //nolint:exhaustruct
		CustomHeader: make(http.Header),
		StatusCode:   http.StatusOK,
	}
	s.primary.ServeHTTP(&rw, r)
	if rw.StatusCode == http.StatusNotFound {
		s.secondary.ServeHTTP(w, r)
		return
	}
	for k, v := range rw.CustomHeader {
		w.Header()[k] = v
	}
	w.WriteHeader(rw.StatusCode)
	_, err := w.Write(rw.Body.Bytes())
	if err != nil {
		log.Println(err)
	}
}
