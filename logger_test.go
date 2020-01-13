package mux

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func testHandler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

func TestLogger(t *testing.T) {
	buffer := new(bytes.Buffer)
	r := NewRouter()
	r.Use(LoggerWithConfig(LogConfig{Output: buffer}))
	r.HandleFunc("/example", testHandler(http.StatusOK)).
		Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")

	rw := NewRecorder()

	req := newRequest("GET", "/example")
	r.ServeHTTP(rw, req)
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	req = newRequest("POST", "/example")
	r.ServeHTTP(rw, req)
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "POST")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	req = newRequest("PUT", "/example")
	r.ServeHTTP(rw, req)
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PUT")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	req = newRequest("DELETE", "/example")
	r.ServeHTTP(rw, req)
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "DELETE")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	req = newRequest("OPTIONS", "/example")
	r.ServeHTTP(rw, req)
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")
}
