package mux

import (
	"net/http"
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
