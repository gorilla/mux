package mux

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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

	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("adds %d middlewares", i+1), func(t *testing.T) {
			mw := &testMiddleware{}
			router.useInterface(mw)
			if len(router.middlewares) != i+1 || router.middlewares[i] != mw {
				t.Fatalf("Middleware %d was not added correctly", i+1)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/", dummyHandler).Methods("GET")

	mw := &testMiddleware{}
	router.useInterface(mw)

	rw := NewRecorder()
	req := newRequest("GET", "/")

	t.Run("regular middleware call", func(t *testing.T) {
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 1 {
			t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
		}
	})

	t.Run("not called for 404", func(t *testing.T) {
		req = newRequest("GET", "/not/found")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 1 {
			t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
		}
	})

	t.Run("not called for method mismatch", func(t *testing.T) {
		req = newRequest("POST", "/")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 1 {
			t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
		}
	})

	t.Run("regular call using function middleware", func(t *testing.T) {
		router.Use(mw.Middleware)
		req = newRequest("GET", "/")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 3 {
			t.Fatalf("Expected %d calls, but got only %d", 3, mw.timesCalled)
		}
	})
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

	t.Run("not called for route outside subrouter", func(t *testing.T) {
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 0 {
			t.Fatalf("Expected %d calls, but got only %d", 0, mw.timesCalled)
		}
	})

	t.Run("not called for subrouter root 404", func(t *testing.T) {
		req = newRequest("GET", "/sub/")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 0 {
			t.Fatalf("Expected %d calls, but got only %d", 0, mw.timesCalled)
		}
	})

	t.Run("called once for route inside subrouter", func(t *testing.T) {
		req = newRequest("GET", "/sub/x")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 1 {
			t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
		}
	})

	t.Run("not called for 404 inside subrouter", func(t *testing.T) {
		req = newRequest("GET", "/sub/not/found")
		router.ServeHTTP(rw, req)
		if mw.timesCalled != 1 {
			t.Fatalf("Expected %d calls, but got only %d", 1, mw.timesCalled)
		}
	})

	t.Run("middleware added to router", func(t *testing.T) {
		router.useInterface(mw)

		t.Run("called once for route outside subrouter", func(t *testing.T) {
			req = newRequest("GET", "/")
			router.ServeHTTP(rw, req)
			if mw.timesCalled != 2 {
				t.Fatalf("Expected %d calls, but got only %d", 2, mw.timesCalled)
			}
		})

		t.Run("called twice for route inside subrouter", func(t *testing.T) {
			req = newRequest("GET", "/sub/x")
			router.ServeHTTP(rw, req)
			if mw.timesCalled != 4 {
				t.Fatalf("Expected %d calls, but got only %d", 4, mw.timesCalled)
			}
		})
	})
}

func TestMiddlewareExecution(t *testing.T) {
	mwStr := []byte("Middleware\n")
	handlerStr := []byte("Logic\n")

	router := NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, e *http.Request) {
		w.Write(handlerStr)
	})

	t.Run("responds normally without middleware", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/")

		router.ServeHTTP(rw, req)

		if !bytes.Equal(rw.Body.Bytes(), handlerStr) {
			t.Fatal("Handler response is not what it should be")
		}
	})

	t.Run("responds with handler and middleware response", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/")

		router.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(mwStr)
				h.ServeHTTP(w, r)
			})
		})

		router.ServeHTTP(rw, req)
		if !bytes.Equal(rw.Body.Bytes(), append(mwStr, handlerStr...)) {
			t.Fatal("Middleware + handler response is not what it should be")
		}
	})
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
	t.Run("not called", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/notfound")

		router.ServeHTTP(rw, req)
		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a 404")
		}
	})

	t.Run("not called with custom not found handler", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/notfound")

		router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("Custom 404 handler"))
		})
		router.ServeHTTP(rw, req)

		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a custom 404")
		}
	})
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

	t.Run("not called", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("POST", "/")

		router.ServeHTTP(rw, req)
		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a method mismatch")
		}
	})

	t.Run("not called with custom method not allowed handler", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("POST", "/")

		router.MethodNotAllowedHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("Method not allowed"))
		})
		router.ServeHTTP(rw, req)

		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a method mismatch")
		}
	})
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

	t.Run("not called", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/sub/notfound")

		router.ServeHTTP(rw, req)
		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a 404")
		}
	})

	t.Run("not called with custom not found handler", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/sub/notfound")

		subrouter.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("Custom 404 handler"))
		})
		router.ServeHTTP(rw, req)

		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a custom 404")
		}
	})
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

	t.Run("not called", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("POST", "/sub/")

		router.ServeHTTP(rw, req)
		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a method mismatch")
		}
	})

	t.Run("not called with custom method not allowed handler", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("POST", "/sub/")

		router.MethodNotAllowedHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("Method not allowed"))
		})
		router.ServeHTTP(rw, req)

		if bytes.Contains(rw.Body.Bytes(), mwStr) {
			t.Fatal("Middleware was called for a method mismatch")
		}
	})
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
		{"/g/{o}", "a", "POST", "/g/asdf", "POST,PUT,GET,OPTIONS"},
		{"/g/{o}", "b", "PUT", "/g/bla", "POST,PUT,GET,OPTIONS"},
		{"/g/{o}", "c", "GET", "/g/orilla", "POST,PUT,GET,OPTIONS"},
		{"/g", "d", "POST", "/g", "POST,OPTIONS"},
	}

	for _, tt := range cases {
		router.HandleFunc(tt.path, stringHandler(tt.response)).Methods(tt.method)
	}

	router.Use(CORSMethodMiddleware(router))

	for i, tt := range cases {
		t.Run(fmt.Sprintf("cases[%d]", i), func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := newRequest(tt.method, tt.testURL)

			router.ServeHTTP(rr, req)

			if rr.Body.String() != tt.response {
				t.Errorf("Expected body '%s', found '%s'", tt.response, rr.Body.String())
			}

			allowedMethods := rr.Header().Get("Access-Control-Allow-Methods")

			if allowedMethods != tt.expectedAllowedMethods {
				t.Errorf("Expected Access-Control-Allow-Methods '%s', found '%s'", tt.expectedAllowedMethods, allowedMethods)
			}
		})
	}
}

func TestMiddlewareOnMultiSubrouter(t *testing.T) {
	first := "first"
	second := "second"
	notFound := "404 not found"

	router := NewRouter()
	firstSubRouter := router.PathPrefix("/").Subrouter()
	secondSubRouter := router.PathPrefix("/").Subrouter()

	router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(notFound))
	})

	firstSubRouter.HandleFunc("/first", func(w http.ResponseWriter, r *http.Request) {

	})

	secondSubRouter.HandleFunc("/second", func(w http.ResponseWriter, r *http.Request) {

	})

	firstSubRouter.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(first))
			h.ServeHTTP(w, r)
		})
	})

	secondSubRouter.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(second))
			h.ServeHTTP(w, r)
		})
	})

	t.Run("/first uses first middleware", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/first")

		router.ServeHTTP(rw, req)
		if rw.Body.String() != first {
			t.Fatalf("Middleware did not run: expected %s middleware to write a response (got %s)", first, rw.Body.String())
		}
	})

	t.Run("/second uses second middleware", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/second")

		router.ServeHTTP(rw, req)
		if rw.Body.String() != second {
			t.Fatalf("Middleware did not run: expected %s middleware to write a response (got %s)", second, rw.Body.String())
		}
	})

	t.Run("uses not found handler", func(t *testing.T) {
		rw := NewRecorder()
		req := newRequest("GET", "/second/not-exist")

		router.ServeHTTP(rw, req)
		if rw.Body.String() != notFound {
			t.Fatalf("Notfound handler did not run: expected %s for not-exist, (got %s)", notFound, rw.Body.String())
		}
	})
}
