//go:build go1.9
// +build go1.9

package mux

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// FilepathAbsServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
// This is a negative test case where `filepath.Abs` will return path value like `D:\`
// if our route is `/`. As per docs: Abs returns an absolute representation of path.
// If the path is not absolute it will be joined with the current working directory to turn
// it into an absolute path. The absolute path name for a given file is not guaranteed to
// be unique. Abs calls Clean on the result.
func (h spaHandler) FilepathAbsServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	// internally calls path.Clean path to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check whether a file exists at the given path
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func TestSchemeMatchers(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("hello http world"))
	}).Schemes("http")
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("hello https world"))
	}).Schemes("https")

	assertResponseBody := func(t *testing.T, s *httptest.Server, expectedBody string) {
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
		if !bytes.Equal(body, []byte(expectedBody)) {
			t.Fatalf("response should be hello world, was: %q", string(body))
		}
	}

	t.Run("httpServer", func(t *testing.T) {
		s := httptest.NewServer(router)
		defer s.Close()
		assertResponseBody(t, s, "hello http world")
	})
	t.Run("httpsServer", func(t *testing.T) {
		s := httptest.NewTLSServer(router)
		defer s.Close()
		assertResponseBody(t, s, "hello https world")
	})
}

func TestServeHttpFilepathAbs(t *testing.T) {
	// create a diretory name `build`
	os.Mkdir("build", 0700)

	// create a file `index.html` and write below content
	htmlContent := []byte(`<html><head><title>hello</title></head><body>world</body></html>`)
	err := os.WriteFile("./build/index.html", htmlContent, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// make new request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	spa := spaHandler{staticPath: "./build", indexPath: "index.html"}
	spa.FilepathAbsServeHTTP(rr, req)

	status := rr.Code
	if runtime.GOOS != "windows" && status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	} else if runtime.GOOS == "windows" && rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code in case of windows: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if runtime.GOOS != "windows" && rr.Body.String() != string(htmlContent) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), string(htmlContent))
	} else if runtime.GOOS == "windows" && !strings.Contains(rr.Body.String(), "syntax is incorrect.") {
		t.Errorf("handler returned unexpected body in case of windows: got %v want %v",
			rr.Body.String(), string(htmlContent))
	}
}

func TestServeHttpFilepathJoin(t *testing.T) {
	// create a diretory name `build`
	os.Mkdir("build", 0700)

	// create a file `index.html` and write below content
	htmlContent := []byte(`<html><head><title>hello</title></head><body>world</body></html>`)
	err := os.WriteFile("./build/index.html", htmlContent, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// make new request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	spa := spaHandler{staticPath: "./build", indexPath: "index.html"}
	spa.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if rr.Body.String() != string(htmlContent) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), string(htmlContent))
	}
}
