package mux

import (
	"errors"
	"net/http"
	"reflect"
	"regexp"
	"sync"
	"testing"
)

var testNewRouterMu sync.Mutex
var testHandler = http.NotFoundHandler()

func BenchmarkNewRouter(b *testing.B) {
	testNewRouterMu.Lock()
	defer testNewRouterMu.Unlock()

	// Set the RegexpCompileFunc to the default regexp.Compile.
	RegexpCompileFunc = regexp.Compile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testNewRouter(b, testHandler)
	}
}

func BenchmarkNewRouterRegexpFunc(b *testing.B) {
	testNewRouterMu.Lock()
	defer testNewRouterMu.Unlock()

	// We preallocate the size to 8.
	cache := make(map[string]*regexp.Regexp, 8)

	// Override the RegexpCompileFunc to reuse compiled expressions
	// from the `cache` map. Real world caches should have eviction
	// policies or some sort of approach to limit memory use.
	RegexpCompileFunc = func(expr string) (*regexp.Regexp, error) {
		if regex, ok := cache[expr]; ok {
			return regex, nil
		}

		regex, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}

		cache[expr] = regex
		return regex, nil
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testNewRouter(b, testHandler)
	}
}

func testNewRouter(_ testing.TB, handler http.Handler) {
	r := NewRouter()
	// A route with a route variable:
	r.Handle("/metrics/{type}", handler)
	r.Queries("orgID", "{orgID:[0-9]*?}")
	r.Host("{subdomain}.domain.com")
}

func TestRouteMetadata(t *testing.T) {
	router := NewRouter()
	rw := NewRecorder()

	expectedMap := make(map[any]any)
	expectedMap["key"] = "value"

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		route := CurrentRoute(r)
		metadata := route.GetMetadata()

		if !reflect.DeepEqual(metadata, expectedMap) {
			println(metadata)
			t.Fatalf("Expected map does not equal the metadata map")
		}

	}).Metadata("key", "value")

	router.HandleFunc("/single-value", func(w http.ResponseWriter, r *http.Request) {
		route := CurrentRoute(r)
		value, err := route.GetMetadataValue("key")
		if err != nil {
			t.Fatalf("Expected metadata value to be present, but gave error: %s", err)
		}

		stringValue, ok := value.(string)
		if !ok {
			t.Fatalf("Expected metadata value to be string, but was: %s", reflect.TypeOf(value))
		}

		if stringValue != "value" {
			t.Fatalf("Expected metadata value to be '%s', but got '%s'", "value", stringValue)
		}

		_, err = route.GetMetadataValue("key2")
		if err == nil {
			t.Fatalf("Expected metadata key not to be present and error, but error was nil")
		}

		if !errors.Is(err, ErrMetadataKeyNotFound) {
			t.Fatalf("Expected error to be ErrMetadataKeyNotFound but got: %s", err)
		}

	}).Metadata("key", "value")

	router.HandleFunc("/single-value-fallback", func(w http.ResponseWriter, r *http.Request) {
		route := CurrentRoute(r)
		value := route.GetMetadataValueOr("key", "value-fallback")

		stringValue, ok := value.(string)
		if !ok {
			t.Fatalf("Expected metadata value to be string, but was: %s", reflect.TypeOf(value))
		}

		if stringValue != "value" {
			t.Fatalf("Expected metadata value to be '%s', but got '%s'", "value", stringValue)
		}

		fallbackValue := route.GetMetadataValueOr("key2", "value2")
		fallbackStringValue, ok := fallbackValue.(string)
		if !ok {
			t.Fatalf("Expected metadata value to be string, but was: %s", reflect.TypeOf(value))
		}

		if fallbackStringValue != "value2" {
			t.Fatalf("Expected metadata value to be '%s', but got '%s'", "value2", fallbackStringValue)
		}

	}).Metadata("key", "value")

	t.Run("get metadata map", func(t *testing.T) {
		req := newRequest("GET", "/")
		router.ServeHTTP(rw, req)
	})

	t.Run("get metadata value", func(t *testing.T) {
		req := newRequest("GET", "/single-value")
		router.ServeHTTP(rw, req)
	})

	t.Run("get metadata value or fallback", func(t *testing.T) {
		req := newRequest("GET", "/single-value-fallback")
		router.ServeHTTP(rw, req)
	})
}
