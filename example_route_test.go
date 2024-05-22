package mux_test

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// ExampleRouteHeadersRegexp demonstrates setting a regular expression matcher for
// the header value. A plain word will match any value that contains a
// matching substring as if the pattern was wrapped with `.*`.
func ExampleRoute_HeadersRegexp() {
	r := mux.NewRouter()
	route := r.NewRoute().HeadersRegexp("Accept", "html")

	req1, err := http.NewRequest("GET", "example.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req1.Header.Add("Accept", "text/plain")
	req1.Header.Add("Accept", "text/html")

	req2, err := http.NewRequest("GET", "example.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req2.Header.Set("Accept", "application/xhtml+xml")

	matchInfo := &mux.RouteMatch{}
	fmt.Printf("Match: %v %q\n", route.Match(req1, matchInfo), req1.Header["Accept"])
	fmt.Printf("Match: %v %q\n", route.Match(req2, matchInfo), req2.Header["Accept"])
	// Output:
	// Match: true ["text/plain" "text/html"]
	// Match: true ["application/xhtml+xml"]
}

// ExampleRouteHeadersRegexpExactMatch demonstrates setting a strict regular expression matcher
// for the header value. Using the start and end of string anchors, the
// value must be an exact match.
func ExampleRoute_HeadersRegexp_exactMatch() {
	r := mux.NewRouter()
	route := r.NewRoute().HeadersRegexp("Origin", "^https://example.co$")

	yes, err := http.NewRequest("GET", "example.co", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	yes.Header.Set("Origin", "https://example.co")

	no, err := http.NewRequest("GET", "example.co.uk", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	no.Header.Set("Origin", "https://example.co.uk")

	matchInfo := &mux.RouteMatch{}
	fmt.Printf("Match: %v %q\n", route.Match(yes, matchInfo), yes.Header["Origin"])
	fmt.Printf("Match: %v %q\n", route.Match(no, matchInfo), no.Header["Origin"])
	// Output:
	// Match: true ["https://example.co"]
	// Match: false ["https://example.co.uk"]
}
