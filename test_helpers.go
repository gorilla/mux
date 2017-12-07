// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import "net/http"

// SetURLVars sets URL variables. This can be used to simplify the testing of
// request handlers.
// Alternatively, URL variables can be set by making a route that captures the
// required variables, starting a server and sending the request to that
// server.
func SetURLVars(r *http.Request, val map[string]string) *http.Request {
	return setVars(r, val)
}
