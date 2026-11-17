// Copyright 2026 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestURLPathLiteralPercent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		template string
		pairs    []string
		want     string
	}{
		{name: "static percent", template: "/100%", want: "/100%"},
		{name: "static format verb", template: "/%s", want: "/%s"},
		{name: "static escaped percent", template: "/%%", want: "/%%"},
		{name: "before variable", template: "/%s/{id}", pairs: []string{"id", "42"}, want: "/%s/42"},
		{name: "after variable", template: "/{id}/%s", pairs: []string{"id", "42"}, want: "/42/%s"},
		{name: "between variables", template: "/{first}/%2F/{second}", pairs: []string{"first", "one", "second", "two"}, want: "/one/%2F/two"},
		{name: "percent in variable", template: "/{id}", pairs: []string{"id", "100%"}, want: "/100%"},
		{name: "plain static", template: "/plain", want: "/plain"},
		{name: "plain variable", template: "/{id}", pairs: []string{"id", "42"}, want: "/42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			route := NewRouter().Path(tc.template)
			for _, build := range []struct {
				name string
				fn   func(...string) (*url.URL, error)
			}{{"URL", route.URL}, {"URLPath", route.URLPath}} {
				t.Run(build.name, func(t *testing.T) {
					built, err := build.fn(tc.pairs...)
					if err != nil {
						t.Fatal(err)
					}
					if built.Path != tc.want {
						t.Errorf("built path = %q, want %q", built.Path, tc.want)
					}
					request := httptest.NewRequest("GET", built.String(), nil)
					if !route.Match(request, &RouteMatch{}) {
						t.Errorf("built URL %q does not match route %q", built.String(), tc.template)
					}
				})
			}
		})
	}
}

func TestRouteRegexpURLLiteralPercent(t *testing.T) {
	for _, typ := range []struct {
		name string
		kind regexpType
	}{
		{"path", regexpTypePath},
		{"host", regexpTypeHost},
		{"query", regexpTypeQuery},
	} {
		for _, tc := range []struct {
			name     string
			template string
			value    string
			want     string
		}{
			{"static", "prefix=%s", "", "prefix=%s"},
			{"before", "prefix=%s{value}", "word", "prefix=%sword"},
			{"after", "prefix={value}%s", "word", "prefix=word%s"},
			{"variable", "prefix={value}", "100%", "prefix=100%"},
			{"plain", "prefix={value}", "word", "prefix=word"},
		} {
			t.Run(typ.name+"/"+tc.name, func(t *testing.T) {
				r, err := newRouteRegexp(tc.template, typ.kind, routeRegexpOptions{})
				if err != nil {
					t.Fatal(err)
				}
				want := tc.want
				if typ.kind == regexpTypeQuery && tc.name == "variable" {
					want = "prefix=100%25"
				}
				got, err := r.url(map[string]string{"value": tc.value})
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("url = %q, want %q", got, want)
				}
			})
		}
	}
}
