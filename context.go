// +build go1.7

package mux

import (
	"context"
	"net/http"
)

func init() {
	contextSet = func(r *http.Request, key, val interface{}) *http.Request {
		return r.WithContext(context.WithValue(r.Context(), key, val))
	}
	contextGet = func(r *http.Request, key interface{}) interface{} {
		return r.Context().Value(key)
	}
}
