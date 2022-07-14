// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMux(b *testing.B) {
	router := new(Router)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/v1/{v1}", handler)

	request, _ := http.NewRequest("GET", "/v1/anything", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

// Benchmark Go's built in serve mux
func BenchmarkMuxGoStd(b *testing.B) {
	router := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/foo", handler)
	router.HandleFunc("/bar", handler)
	router.HandleFunc("/fizz", handler)
	router.HandleFunc("/buzz", handler)

	request, _ := http.NewRequest("GET", "/buzz", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

func BenchmarkMuxGorilla(b *testing.B) {
	router := new(Router)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/foo", handler)
	router.HandleFunc("/bar", handler)
	router.HandleFunc("/fizz", handler)
	router.HandleFunc("/buzz", handler)

	request, _ := http.NewRequest("GET", "/buzz", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

func BenchmarkMuxGorillaSkipClean(b *testing.B) {
	router := new(Router)
	router.SkipClean(true)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/foo", handler)
	router.HandleFunc("/bar", handler)
	router.HandleFunc("/fizz", handler)
	router.HandleFunc("/buzz", handler)

	request, _ := http.NewRequest("GET", "/buzz", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

func BenchmarkMux5Middleware(b *testing.B) {
	router := new(Router)
	router.SkipClean(true)
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	handler := func(w http.ResponseWriter, r *http.Request) {}
	for i := 0; i <= 5; i++ {
		router.Use(middleware)
	}
	router.HandleFunc("/foo", handler)
	router.HandleFunc("/bar", handler)
	router.HandleFunc("/fizz", handler)
	router.HandleFunc("/buzz", handler)

	request, _ := http.NewRequest("GET", "/buzz", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

func BenchmarkMux20Middleware(b *testing.B) {
	router := new(Router)
	router.SkipClean(true)
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	handler := func(w http.ResponseWriter, r *http.Request) {}
	for i := 0; i <= 20; i++ {
		router.Use(middleware)
	}
	router.HandleFunc("/foo", handler)
	router.HandleFunc("/bar", handler)
	router.HandleFunc("/fizz", handler)
	router.HandleFunc("/buzz", handler)

	request, _ := http.NewRequest("GET", "/buzz", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, request)
	}
}

func BenchmarkMuxAlternativeInRegexp(b *testing.B) {
	router := new(Router)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/v1/{v1:(?:a|b)}", handler)

	requestA, _ := http.NewRequest("GET", "/v1/a", nil)
	requestB, _ := http.NewRequest("GET", "/v1/b", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, requestA)
		router.ServeHTTP(nil, requestB)
	}
}

func BenchmarkManyPathVariables(b *testing.B) {
	router := new(Router)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	router.HandleFunc("/v1/{v1}/{v2}/{v3}/{v4}/{v5}", handler)

	matchingRequest, _ := http.NewRequest("GET", "/v1/1/2/3/4/5", nil)
	notMatchingRequest, _ := http.NewRequest("GET", "/v1/1/2/3/4", nil)
	recorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, matchingRequest)
		router.ServeHTTP(recorder, notMatchingRequest)
	}
}

func BenchmarkManyPathVariablesLong(b *testing.B) {
	router := new(Router)
	handler := func(w http.ResponseWriter, r *http.Request) {}
	lorem := "Lorem_ipsum_dolor_sit_amet,_consectetur_adipiscing_elit_sed_do_eiusmod_tempor_incididunt_ut_labore_et_dolore_magna_aliqua"
	router.HandleFunc("/v1/{v1}/{v2}/{v3}/{v4}/{v5}", handler)

	matchingRequest, _ := http.NewRequest("GET", fmt.Sprintf("/v1/%[1]s/%[1]s/%[1]s/%[1]s/%[1]s", lorem), nil)
	notMatchingRequest, _ := http.NewRequest("GET", fmt.Sprintf("/v1/%[1]s/%[1]s/%[1]s/%[1]s", lorem), nil)
	recorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(nil, matchingRequest)
		router.ServeHTTP(recorder, notMatchingRequest)
	}
}
