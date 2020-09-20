package mux_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// This example demonstrates alias pattern registration on router
func ExampleRouter_RegisterPattern() {

	r := mux.NewRouter().RegisterPattern("uuid", "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}")
	route := r.Path("/category/{id:uuid}")

	yes, _ := http.NewRequest("GET", "example.co/category/abe193ed-e0bc-4e1b-8e3c-736d5b381b60", nil)
	no, _ := http.NewRequest("GET", "example.co/category/42", nil)

	mathInfo := &mux.RouteMatch{}
	fmt.Printf("Match: %v %q\n", route.Match(yes, mathInfo), yes.URL.Path)
	fmt.Printf("Match: %v %q\n", route.Match(no, mathInfo), no.URL.Path)

	// Output
	// Match: true /category/abe193ed-e0bc-4e1b-8e3c-736d5b381b60
	// Match: false /category/42
}