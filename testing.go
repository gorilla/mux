package mux

import "net/http"

// TestSetURLParam set url params
func TestSetURLParam(r *http.Request, val map[string]string) *http.Request {
	return setVars(r, val)
}
