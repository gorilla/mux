// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRoute(t *testing.T) {
	var route *Route
	var request *http.Request
	var vars map[string]string
	var host, path, url string

	// Setup an id so we can see which test failed. :)
	var idValue int
	id := func() int {
		idValue++
		return idValue
	}

	// Host -------------------------------------------------------------------

	route = new(Route).Host("aaa.bbb.ccc")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{}
	host = "aaa.bbb.ccc"
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Host("aaa.{v1:[a-z]{3}}.ccc")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{"v1": "bbb"}
	host = "aaa.bbb.ccc"
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Host("{v1:[a-z]{3}}.{v2:[a-z]{3}}.{v3:[a-z]{3}}")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{"v1": "aaa", "v2": "bbb", "v3": "ccc"}
	host = "aaa.bbb.ccc"
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Path -------------------------------------------------------------------

	route = new(Route).Path("/111/222/333")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{}
	host = ""
	path = "/111/222/333"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/1/2/3", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Path("/111/{v1:[0-9]{3}}/333")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{"v1": "222"}
	host = ""
	path = "/111/222/333"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/111/aaa/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Path("/{v1:[0-9]{3}}/{v2:[0-9]{3}}/{v3:[0-9]{3}}")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{"v1": "111", "v2": "222", "v3": "333"}
	host = ""
	path = "/111/222/333"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/111/aaa/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// PathPrefix -------------------------------------------------------------

	route = new(Route).PathPrefix("/111")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{}
	host = ""
	path = "/111"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/1/2/3", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).PathPrefix("/111/{v1:[0-9]{3}}")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{"v1": "222"}
	host = ""
	path = "/111/222"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/111/aaa/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).PathPrefix("/{v1:[0-9]{3}}/{v2:[0-9]{3}}")
	request, _ = http.NewRequest("GET", "http://localhost/111/222/333", nil)
	vars = map[string]string{"v1": "111", "v2": "222"}
	host = ""
	path = "/111/222"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/111/aaa/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Host + Path ------------------------------------------------------------

	route = new(Route).Host("aaa.bbb.ccc").Path("/111/222/333")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Host("aaa.{v1:[a-z]{3}}.ccc").Path("/111/{v2:[0-9]{3}}/333")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{"v1": "bbb", "v2": "222"}
	host = "aaa.bbb.ccc"
	path = "/111/222/333"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	route = new(Route).Host("{v1:[a-z]{3}}.{v2:[a-z]{3}}.{v3:[a-z]{3}}").Path("/{v4:[0-9]{3}}/{v5:[0-9]{3}}/{v6:[0-9]{3}}")
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc/111/222/333", nil)
	vars = map[string]string{"v1": "aaa", "v2": "bbb", "v3": "ccc", "v4": "111", "v5": "222", "v6": "333"}
	host = "aaa.bbb.ccc"
	path = "/111/222/333"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.222.ccc/111/222/333", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Headers ----------------------------------------------------------------

	route = new(Route).Headers("foo", "bar", "baz", "ding")
	request, _ = http.NewRequest("GET", "http://localhost", nil)
	request.Header.Add("foo", "bar")
	request.Header.Add("baz", "ding")
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost", nil)
	request.Header.Add("foo", "bar")
	request.Header.Add("baz", "dong")
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Methods ----------------------------------------------------------------

	route = new(Route).Methods("GET", "POST")
	request, _ = http.NewRequest("GET", "http://localhost", nil)
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	request, _ = http.NewRequest("POST", "http://localhost", nil)
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("PUT", "http://localhost", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Queries ----------------------------------------------------------------

	route = new(Route).Queries("foo", "bar", "baz", "ding")
	request, _ = http.NewRequest("GET", "http://localhost?foo=bar&baz=ding", nil)
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost?foo=bar&baz=dong", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Schemes ----------------------------------------------------------------

	route = new(Route).Schemes("https", "ftp")
	request, _ = http.NewRequest("GET", "https://localhost", nil)
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	request, _ = http.NewRequest("GET", "ftp://localhost", nil)
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// Custom -----------------------------------------------------------------

	m := func(r *http.Request, m *RouteMatch) bool {
		if r.URL.Host == "aaa.bbb.ccc" {
			return true
		}
		return false
	}
	route = new(Route).MatcherFunc(m)
	request, _ = http.NewRequest("GET", "http://aaa.bbb.ccc", nil)
	vars = map[string]string{}
	host = ""
	path = ""
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://aaa.ccc.bbb", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)
}

func TestSubRouter(t *testing.T) {
	var route *Route
	var request *http.Request
	var vars map[string]string
	var host, path, url string

	subrouter := new(Route).Host("{v1:[a-z]+}.google.com").Subrouter()

	// Setup an id so we can see which test failed. :)
	var idValue int
	id := func() int {
		idValue++
		return idValue
	}

	// ------------------------------------------------------------------------

	route = subrouter.Path("/{v2:[a-z]+}")
	request, _ = http.NewRequest("GET", "http://aaa.google.com/bbb", nil)
	vars = map[string]string{"v1": "aaa", "v2": "bbb"}
	host = "aaa.google.com"
	path = "/bbb"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://111.google.com/111", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)

	// ------------------------------------------------------------------------

	subrouter = new(Route).PathPrefix("/foo/{v1}").Subrouter()
	route = subrouter.Path("/baz/{v2}")
	request, _ = http.NewRequest("GET", "http://localhost/foo/bar/baz/ding", nil)
	vars = map[string]string{"v1": "bar", "v2": "ding"}
	host = ""
	path = "/foo/bar/baz/ding"
	url = host + path
	testRoute(t, id(), true, route, request, vars, host, path, url)
	// Non-match for the same config.
	request, _ = http.NewRequest("GET", "http://localhost/foo/bar", nil)
	testRoute(t, id(), false, route, request, vars, host, path, url)
}

func TestNamedRoutes(t *testing.T) {
	r1 := NewRouter()
	r1.NewRoute().Name("a")
	r1.NewRoute().Name("b")
	r1.NewRoute().Name("c")

	r2 := r1.NewRoute().Subrouter()
	r2.NewRoute().Name("d")
	r2.NewRoute().Name("e")
	r2.NewRoute().Name("f")

	r3 := r2.NewRoute().Subrouter()
	r3.NewRoute().Name("g")
	r3.NewRoute().Name("h")
	r3.NewRoute().Name("i")

	if r1.namedRoutes == nil || len(r1.namedRoutes) != 9 {
		t.Errorf("Expected 9 named routes, got %v", r1.namedRoutes)
	} else if r1.Get("i") == nil {
		t.Errorf("Subroute name not registered")
	}
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func getRouteTemplate(route *Route) string {
	host, path := "none", "none"
	if route.regexp != nil {
		if route.regexp.host != nil {
			host = route.regexp.host.template
		}
		if route.regexp.path != nil {
			path = route.regexp.path.template
		}
	}
	return fmt.Sprintf("Host: %v, Path: %v", host, path)
}

func testRoute(t *testing.T, id int, shouldMatch bool, route *Route,
	request *http.Request, vars map[string]string, host, path, url string) {
	var match RouteMatch
	ok := route.Match(request, &match)
	if ok != shouldMatch {
		msg := "Should match"
		if !shouldMatch {
			msg = "Should not match"
		}
		t.Errorf("(%v) %v:\nRoute: %#v\nRequest: %#v\nVars: %v\n", id, msg, route, request, vars)
		return
	}
	if shouldMatch {
		if vars != nil && !stringMapEqual(vars, match.Vars) {
			t.Errorf("(%v) Vars not equal: expected %v, got %v", id, vars, match.Vars)
			return
		}
		if host != "" {
			u, _ := route.URLHost(mapToPairs(match.Vars)...)
			if host != u.Host {
				t.Errorf("(%v) URLHost not equal: expected %v, got %v -- %v", id, host, u.Host, getRouteTemplate(route))
				return
			}
		}
		if path != "" {
			u, _ := route.URLPath(mapToPairs(match.Vars)...)
			if path != u.Path {
				t.Errorf("(%v) URLPath not equal: expected %v, got %v -- %v", id, path, u.Path, getRouteTemplate(route))
				return
			}
		}
		if url != "" {
			u, _ := route.URL(mapToPairs(match.Vars)...)
			if url != u.Host+u.Path {
				t.Errorf("(%v) URL not equal: expected %v, got %v -- %v", id, url, u.Host+u.Path, getRouteTemplate(route))
				return
			}
		}
	}
}

func TestStrictSlash(t *testing.T) {
	var r *Router
	var req *http.Request
	var route *Route
	var match *RouteMatch
	var matched bool

	// StrictSlash should be ignored for path prefix.
	// So we register a route ending in slash but it doesn't attempt to add
	// the slash for a path not ending in slash.
	r = NewRouter()
	r.StrictSlash(true)
	route = r.NewRoute().PathPrefix("/static/")
	req, _ = http.NewRequest("GET", "http://localhost/static/logo.png", nil)
	match = new(RouteMatch)
	matched = r.Match(req, match)
	if !matched {
		t.Errorf("Should match request %q -- %v", req.URL.Path, getRouteTemplate(route))
	}
	if match.Handler != nil {
		t.Errorf("Should not redirect")
	}
}

func mapToPairs(m map[string]string) []string {
	var i int
	p := make([]string, len(m)*2)
	for k, v := range m {
		p[i] = k
		p[i+1] = v
		i += 2
	}
	return p
}

func stringMapEqual(m1, m2 map[string]string) bool {
	nil1 := m1 == nil
	nil2 := m2 == nil
	if nil1 != nil2 || len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if v != m2[k] {
			return false
		}
	}
	return true
}
