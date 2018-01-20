// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import "net/http"

// SetURLVars sets the URL variables for the given request, to be accessed via
// mux.Vars for testing route behaviour. A shallow copy of the *http.Request is
// returned, the given *http.Request is not modified, so you will need to set
// the request to the returned *http.Request like so:
//
//     req, _ := http.NewRequest("GET", "/foo", nil)
//     req = mux.SetURLVars(req, map[string]string{"foo": "bar"})
//
// This API should only be used for testing purposes; it provides a way to
// inject variables into the request context. Alternatively, URL variables
// can be set by making a route that captures the required variables,
// starting a server and sending the request to that server.
func SetURLVars(r *http.Request, val map[string]string) *http.Request {
	return setVars(r, val)
}
