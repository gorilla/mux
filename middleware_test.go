package mux

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testMiddleware struct {
	timesCalled uint
}

func (tm *testMiddleware) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tm.timesCalled++
		h.ServeHTTP(w, r)
	})
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func TestMiddlewareAdd(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/", dummyHandler).Methods("GET")

	mw := &testMiddleware{}

	router.useInterface(mw)
	if len(router.middlewares) != 1 || router.middlewares[0] != mw {
		t.Fatal("Middleware was not added correctly")
	}

	router.Use(mw.Middleware)
	if len(router.middlewares) != 2 {
		t.Fatal("MiddlewareFunc method was not added correctly")
	}

	banalMw := func(handler http.Handler) http.Handler {
		return handler
	}
	router.Use(banalMw)
	if len(router.middlewares) != 3 {
		t.Fatal("MiddlewareFunc method was not added correctly")
	}
}

func TestMiddleware(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/", dummyHandler).Methods("GET")

	mw := &testMiddleware{}
	router.useInterface(mw)

	rw := NewRecorder()
	req := newRequest("GET", "/")

	// Test regular middleware call
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 1 {
		t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
	}

	// Middleware should not be called for 404
	req = newRequest("GET", "/not/found")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 1 {
		t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
	}

	// Middleware should not be called if there is a method mismatch
	req = newRequest("POST", "/")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 1 {
		t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
	}

	// Add the middleware again as function
	router.Use(mw.Middleware)
	req = newRequest("GET", "/")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 3 {
		t.Fatalf("Expected %d calls, but got only %d", 3, mw.timesCalled)
	}

}

func TestMiddlewareSubrouter(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/", dummyHandler).Methods("GET")

	subrouter := router.PathPrefix("/sub").Subrouter()
	subrouter.HandleFunc("/x", dummyHandler).Methods("GET")

	mw := &testMiddleware{}
	subrouter.useInterface(mw)

	rw := NewRecorder()
	req := newRequest("GET", "/")

	router.ServeHTTP(rw, req)
	if mw.timesCalled != 0 {
		t.Fatalf("Expected %d calls, but got only %d", 0, mw.timesCalled)
	}

	req = newRequest("GET", "/sub/")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 0 {
		t.Fatalf("Expected %d calls, but got only %d", 0, mw.timesCalled)
	}

	req = newRequest("GET", "/sub/x")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 1 {
		t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
	}

	req = newRequest("GET", "/sub/not/found")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 1 {
		t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
	}

	router.useInterface(mw)

	req = newRequest("GET", "/")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 2 {
		t.Fatalf("Expected %d calls, but got only %d", 2, mw.timesCalled)
	}

	req = newRequest("GET", "/sub/x")
	router.ServeHTTP(rw, req)
	if mw.timesCalled != 4 {
		t.Fatalf("Expected %d calls, but got only %d", 4, mw.timesCalled)
	}
}

func TestMiddlewareExecution(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})

	rw := NewRecorder()
	req := newRequest("GET", "/")

	// Test handler-only call
	router.ServeHTTP(rw, req)

	if bytes.Compare(rw.Body.Bytes(), handlerStr) != 0 {
		t.Fatal("Handler response is not what it should be")
	}

	// Test middleware call
	rw = NewRecorder()

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(mwStr)
			h.ServeHTTP(w, r)
		})
	})

	router.ServeHTTP(rw, req)
	if bytes.Compare(rw.Body.Bytes(), append(mwStr, handlerStr...)) != 0 {
		t.Fatal("Middleware + handler response is not what it should be")
	}
}

func TestMiddlewareNotFound(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(mwStr)
			h.ServeHTTP(w, r)
		})
	})

	// Test not found call with default handler
	rw := NewRecorder()
	req := newRequest("GET", "/notfound")

	router.ServeHTTP(rw, req)
	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a 404")
	}

	// Test not found call with custom handler
	rw = NewRecorder()
	req = newRequest("GET", "/notfound")

	router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Custom 404 handler"))
	})
	router.ServeHTTP(rw, req)

	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a custom 404")
	}
}

func TestMiddlewareMethodMismatch(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	}).Methods("GET")

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(mwStr)
			h.ServeHTTP(w, r)
		})
	})

	// Test method mismatch
	rw := NewRecorder()
	req := newRequest("POST", "/")

	router.ServeHTTP(rw, req)
	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a method mismatch")
	}

	// Test not found call
	rw = NewRecorder()
	req = newRequest("POST", "/")

	router.MethodNotAllowedHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Method not allowed"))
	})
	router.ServeHTTP(rw, req)

	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a method mismatch")
	}
}

func TestMiddlewareNotFoundSubrouter(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})

	subrouter := router.PathPrefix("/sub/").Subrouter()
	subrouter.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(mwStr)
			h.ServeHTTP(w, r)
		})
	})

	// Test not found call for default handler
	rw := NewRecorder()
	req := newRequest("GET", "/sub/notfound")

	router.ServeHTTP(rw, req)
	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a 404")
	}

	// Test not found call with custom handler
	rw = NewRecorder()
	req = newRequest("GET", "/sub/notfound")

	subrouter.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Custom 404 handler"))
	})
	router.ServeHTTP(rw, req)

	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a custom 404")
	}
}

func TestMiddlewareMethodMismatchSubrouter(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})

	subrouter := router.PathPrefix("/sub/").Subrouter()
	subrouter.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	}).Methods("GET")

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(mwStr)
			h.ServeHTTP(w, r)
		})
	})

	// Test method mismatch without custom handler
	rw := NewRecorder()
	req := newRequest("POST", "/sub/")

	router.ServeHTTP(rw, req)
	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a method mismatch")
	}

	// Test method mismatch with custom handler
	rw = NewRecorder()
	req = newRequest("POST", "/sub/")

	router.MethodNotAllowedHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Method not allowed"))
	})
	router.ServeHTTP(rw, req)

	if bytes.Contains(rw.Body.Bytes(), mwStr) {
		t.Fatal("Middleware was called for a method mismatch")
	}
}

func TestCORSMethodMiddleware(t *testing.T) {
	router := NewRouter()

	cases := []struct {
		path                   string
		response               string
		method                 string
		testURL                string
		expectedAllowedMethods string
	}{
		{"/g/{o}", "a", "POST", "/g/asdf", "POST,PUT,GET"},
		{"/g/{o}", "b", "PUT", "/g/bla", "POST,PUT,GET"},
		{"/g/{o}", "c", "GET", "/g/orilla", "POST,PUT,GET"},
		{"/g", "d", "POST", "/g", "POST"},
	}

	for _, tt := range cases {
		router.HandleFunc(tt.path, stringHandler(tt.response)).Methods(tt.method)
	}

	router.Use(CORSMethodMiddleware(router))

	for _, tt := range cases {
		rr := httptest.NewRecorder()
		req := newRequest(tt.method, tt.testURL)

		router.ServeHTTP(rr, req)

		if rr.Body.String() != tt.response {
			t.Errorf("Expected body '%s', found '%s'", tt.response, rr.Body.String())
		}

		allowedMethods := rr.HeaderMap.Get("Access-Control-Allow-Methods")

		if allowedMethods != tt.expectedAllowedMethods {
			t.Errorf("Expected Access-Control-Allow-Methods '%s', found '%s'", tt.expectedAllowedMethods, allowedMethods)
		}
	}
}

// A route without a Method filter will not trigger the middleware.
func TestCORSMiddlewareOPTIONSWithoutMethodMatcher(t *testing.T) {
	router := NewRouter()

	handlerStr := "a"

	router.HandleFunc("/g/{o}", stringHandler(handlerStr))

	router.Use(CORSMethodMiddleware(router))

	rr := httptest.NewRecorder()
	req := newRequest("OPTIONS", "/g/asdf")

	router.ServeHTTP(rr, req)

	if want, have := rr.HeaderMap.Get("Access-Control-Allow-Methods"), ""; have != want {
		t.Errorf("Expected Access-Control-Allow-Methods '%s', found '%s'", want, have)
	}

	if string(rr.Body.Bytes()) != handlerStr {
		t.Fatal("Handler response is not what it should be")
	}
}

func TestCORSMiddlewareOPTIONSWithMethodMatcher(t *testing.T) {
	router := NewRouter()

	cases := []struct {
		path                   string
		response               string
		method                 string
		testURL                string
		expectedAllowedMethods string
	}{
		{"/g/{o}", "a", "POST", "/g/asdf", "POST,PUT,GET,OPTIONS"},
		{"/g/{o}", "b", "PUT", "/g/bla", "POST,PUT,GET,OPTIONS"},
		{"/g/{o}", "c", "GET", "/g/orilla", "POST,PUT,GET,OPTIONS"},
		{"/g/{o}", "c", "OPTIONS", "/g/orilla", "POST,PUT,GET,OPTIONS"},
	}

	for _, tt := range cases {
		router.HandleFunc(tt.path, stringHandler(tt.response)).Methods(tt.method)
	}

	router.Use(CORSMethodMiddleware(router))

	for _, tt := range cases {
		rr := httptest.NewRecorder()
		req := newRequest("OPTIONS", tt.testURL)

		router.ServeHTTP(rr, req)

		allowedMethods := rr.HeaderMap.Get("Access-Control-Allow-Methods")

		if want, have := 200, rr.Code; have != want {
			t.Errorf("Expected status code %d, found %d", want, have)
		}

		if allowedMethods != tt.expectedAllowedMethods {
			t.Errorf("Expected Access-Control-Allow-Methods '%s', found '%s'", tt.expectedAllowedMethods, allowedMethods)
		}
	}
}

// Create a router, attach a route with a Methods filter, and apply the middleware.
// The middlewares are only executed if the call matches a route.
// Explicitly enable a Methods filter with "OPTIONS" in order for the middleware
// to handle it.
func ExampleCORSMethodMiddleware() {
	router := NewRouter()
	router.Path("/some-path").HandlerFunc(stringHandler("This endpoint is accessible through CORS")).Methods("OPTIONS", "GET", "PUT", "DELETE")
	router.Use(CORSMethodMiddleware(router))

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// To only enable the CORS header for specific matchers, enable the middleware
// on a subrouter.
func ExampleCORSMethodMiddleware_subrouter() {
	router := NewRouter()
	router.Path("/endpoint_without_cors_header").HandlerFunc(stringHandler("no CORS allowed")).Methods("OPTIONS", "GET", "PUT", "DELETE")

	routesWithCORS := router.NewRoute().Subrouter()
	routesWithCORS.Path("/endpoint_with_cors_header").HandlerFunc(stringHandler("CORS allowed here!")).Methods("OPTIONS", "GET", "PUT", "DELETE")
	routesWithCORS.Use(CORSMethodMiddleware(routesWithCORS))

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
