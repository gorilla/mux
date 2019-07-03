// +build go1.9

package mux

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSchemeMatchers(t *testing.T) {
	httpRouter := NewRouter()
	httpRouter.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("hello world"))
	}).Schemes("http")
	httpsRouter := NewRouter()
	httpsRouter.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("hello world"))
	}).Schemes("https")

	assertHelloWorldResponse := func(t *testing.T, s *httptest.Server) {
		resp, err := s.Client().Get(s.URL)
		if err != nil {
			t.Fatalf("unexpected error getting from server: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected a status code of 200, got %v", resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error reading body: %v", err)
		}
		if !bytes.Equal(body, []byte("hello world")) {
			t.Fatalf("response should be hello world, was: %q", string(body))
		}
	}

	t.Run("httpServer", func(t *testing.T) {
		s := httptest.NewServer(httpRouter)
		defer s.Close()
		assertHelloWorldResponse(t, s)
	})
	t.Run("httpsServer", func(t *testing.T) {
		s := httptest.NewTLSServer(httpsRouter)
		defer s.Close()
		assertHelloWorldResponse(t, s)
	})
}
