package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestURLBuilderAlternationValidation(t *testing.T) {
	router := NewRouter()
	path := router.Path("/items/{value:alpha|beta}")
	host := router.Host("{value:alpha|beta}.example.com")
	query := router.Path("/search").Queries("q", "{value:alpha|beta}")
	grouped := router.Path("/grouped/{value:(?:alpha|beta)}")
	childPath := router.PathPrefix("/parent/{value:alpha|beta}").Subrouter().Path("/child")
	childHost := router.Host("{value:alpha|beta}.example.org").Subrouter().Path("/child")
	builders := []struct {
		name  string
		build func(...string) (*url.URL, error)
	}{
		{"path URL", path.URL},
		{"path URLPath", path.URLPath},
		{"host URL", host.URL},
		{"host URLHost", host.URLHost},
		{"query URL", query.URL},
		{"grouped URL", grouped.URL},
		{"inherited path URL", childPath.URL},
		{"inherited path URLPath", childPath.URLPath},
		{"inherited host URL", childHost.URL},
		{"inherited host URLHost", childHost.URLHost},
	}
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"alpha", false},
		{"beta", false},
		{"alphax", true},
		{"xbeta", true},
		{"alphabeta", true},
		{"xalphay", true},
	}
	for _, builder := range builders {
		t.Run(builder.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.value, func(t *testing.T) {
					built, err := builder.build("value", tc.value)
					if tc.wantErr {
						if err == nil {
							t.Fatalf("building with %q returned %s instead of a validation error", tc.value, built)
						}
						if built != nil {
							t.Fatalf("failed URL build returned a URL: %s", built)
						}
						if !strings.Contains(err.Error(), tc.value) || !strings.Contains(err.Error(), "doesn't match") {
							t.Fatalf("unexpected validation error: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatalf("valid alternative %q was rejected: %v", tc.value, err)
					}
					if built == nil || !strings.Contains(built.String(), tc.value) {
						t.Fatalf("valid alternative missing from generated URL: %v", built)
					}
				})
			}
		})
	}
}

func TestURLBuilderAlternationBuildVars(t *testing.T) {
	route := NewRouter().Path("/items/{value:alpha|beta}").BuildVarsFunc(func(values map[string]string) map[string]string {
		values["value"] = strings.ToLower(values["value"])
		return values
	})
	for _, input := range []string{"ALPHA", "BETA"} {
		t.Run(input, func(t *testing.T) {
			built, err := route.URL("value", input)
			if err != nil {
				t.Fatalf("transformed valid alternative was rejected: %v", err)
			}
			if want := "/items/" + strings.ToLower(input); built.Path != want {
				t.Fatalf("generated path = %q, want %q", built.Path, want)
			}
		})
	}
	for _, input := range []string{"ALPHAX", "XBETA"} {
		t.Run(input, func(t *testing.T) {
			built, err := route.URL("value", input)
			if err == nil {
				t.Fatalf("transformed invalid alternative returned %s without an error", built)
			}
			if !strings.Contains(err.Error(), strings.ToLower(input)) {
				t.Fatalf("error does not identify the transformed value: %v", err)
			}
		})
	}
}

func TestURLBuilderAlternationHTTPRoundTrip(t *testing.T) {
	router := NewRouter()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, Vars(r)["value"]); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}
	path := router.HandleFunc("/items/{value:alpha|beta}", handler)
	host := router.Host("{value:alpha|beta}.example.com").Path("/host").HandlerFunc(handler)
	query := router.HandleFunc("/search", handler).Queries("q", "{value:alpha|beta}")
	server := httptest.NewServer(router)
	defer server.Close()
	for _, route := range []*Route{path, host, query} {
		for _, value := range []string{"alpha", "beta"} {
			built, err := route.URL("value", value)
			if err != nil {
				t.Fatalf("building valid URL: %v", err)
			}
			req, err := http.NewRequest(http.MethodGet, server.URL+built.RequestURI(), nil)
			if err != nil {
				t.Fatalf("creating request: %v", err)
			}
			if built.Host != "" {
				req.Host = built.Host
			}
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatalf("requesting generated URL: %v", err)
			}
			body, readErr := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("reading response: read=%v close=%v", readErr, closeErr)
			}
			if resp.StatusCode != http.StatusOK || string(body) != value {
				t.Fatalf("generated URL %s returned status %d and body %q", built, resp.StatusCode, body)
			}
		}
	}
	for _, target := range []struct {
		path string
		host string
	}{
		{"/items/alphax", ""},
		{"/items/xbeta", ""},
		{"/host", "alphax.example.com"},
		{"/host", "xbeta.example.com"},
		{"/search?q=alphax", ""},
		{"/search?q=xbeta", ""},
	} {
		req, err := http.NewRequest(http.MethodGet, server.URL+target.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if target.host != "" {
			req.Host = target.host
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("invalid alternative at %s (host %s) returned status %d", target.path, target.host, resp.StatusCode)
		}
	}
}
