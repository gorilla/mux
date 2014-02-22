package main

import (
	"github.com/tarrsalah/mux"
	"net/http"
)

const (
	http_port = ":8080"
)

var root *mux.Router

func E1() {
	root = mux.NewRouter()
	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello gopher gorilla !"))
	})
}

func main() {
	E1()
	http.Handle("/", root)
	http.ListenAndServe(http_port, nil)
}
