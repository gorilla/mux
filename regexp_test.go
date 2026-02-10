package mux

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func Test_newRouteRegexp_Errors(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"/{}", `mux: missing name or pattern in "{}"`},
		{"/{:.}", `mux: missing name or pattern in "{:.}"`},
		{"/{a:}", `mux: missing name or pattern in "{a:}"`},
		{"/{id:abc(}", `mux: error compiling regex for "{id:abc(}":`},
	}

	for _, tc := range tests {
		t.Run("Test case for "+tc.in, func(t *testing.T) {
			_, err := newRouteRegexp(tc.in, 0, routeRegexpOptions{})
			if err != nil {
				if strings.HasPrefix(err.Error(), tc.out) {
					return
				}
				t.Errorf("Resulting error does not contain %q as expected, error: %s", tc.out, err)
			} else {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func Test_findFirstQueryKey(t *testing.T) {
	tests := []string{
		"a=1&b=2",
		"a=1&a=2&a=banana",
		"ascii=%3Ckey%3A+0x90%3E",
		"a==",
		"a=%2",
		"a=20&%20%3F&=%23+%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09:%2F@$%27%28%29%2A%2C%3B&a=30",
		"a=1& ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;&a=5",
		"a=xxxxxxxxxxxxxxxx&b=YYYYYYYYYYYYYYY&c=ppppppppppppppppppp&f=ttttttttttttttttt&a=uuuuuuuuuuuuu",
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			// Check against url.ParseQuery, ignoring the error.
			all, _ := url.ParseQuery(query)
			for key, want := range all {
				t.Run(key, func(t *testing.T) {
					got, ok := findFirstQueryKey(query, key)
					if !ok {
						t.Error("Did not get expected key", key)
					}
					if !reflect.DeepEqual(got, want[0]) {
						t.Errorf("findFirstQueryKey(%s,%s) = %v, want %v", query, key, got, want[0])
					}
				})
			}
		})
	}
}

// Test_findFirstQueryKey_semicolonIsNotSeparator verifies that semicolons
// are NOT treated as query parameter separators, in compliance with the
// URL Living Standard and Go's net/url since Go 1.17.
// See https://github.com/gorilla/mux/issues/781
func Test_findFirstQueryKey_semicolonIsNotSeparator(t *testing.T) {
	// "a=1;b=2" should be treated as a single key-value pair where
	// the key is "a" and the value is "1;b=2", not as two separate pairs.
	t.Run("semicolon in value is not a separator", func(t *testing.T) {
		_, ok := findFirstQueryKey("a=1;b=2", "b")
		if ok {
			t.Error("semicolon should not act as a query parameter separator")
		}
		val, ok := findFirstQueryKey("a=1;b=2", "a")
		if !ok {
			t.Error("expected to find key 'a'")
		}
		if val != "1;b=2" {
			t.Errorf("expected value '1;b=2', got %q", val)
		}
	})

	// Verify that the parser differential described in issue #781 is fixed:
	// for "a;id=42&id=1", the key "id" should resolve to "1" (the
	// ampersand-separated part), not "42" (the semicolon-separated part).
	t.Run("no parser differential with net/url", func(t *testing.T) {
		val, ok := findFirstQueryKey("a;id=42&id=1", "id")
		if !ok {
			t.Error("expected to find key 'id'")
		}
		if val != "1" {
			t.Errorf("expected value '1', got %q", val)
		}
		// "a;id=42" should be treated as a single key (not split on semicolon).
		_, ok = findFirstQueryKey("a;id=42&id=1", "a;id")
		if !ok {
			t.Error("expected to find key 'a;id' (semicolon is part of the key)")
		}
	})

	// Additional case: semicolons mixed with ampersands should only split on ampersands.
	t.Run("mixed semicolons and ampersands", func(t *testing.T) {
		val, ok := findFirstQueryKey("foo=foo;bar=bar&baz=ding", "foo")
		if !ok {
			t.Error("expected to find key 'foo'")
		}
		if val != "foo;bar=bar" {
			t.Errorf("expected value 'foo;bar=bar', got %q", val)
		}
		val, ok = findFirstQueryKey("foo=foo;bar=bar&baz=ding", "baz")
		if !ok {
			t.Error("expected to find key 'baz'")
		}
		if val != "ding" {
			t.Errorf("expected value 'ding', got %q", val)
		}
	})
}

func Benchmark_findQueryKey(b *testing.B) {
	tests := []string{
		"a=1&b=2",
		"ascii=%3Ckey%3A+0x90%3E",
		"a=20&%20%3F&=%23+%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09:%2F@$%27%28%29%2A%2C%3B&a=30",
		"a=xxxxxxxxxxxxxxxx&bbb=YYYYYYYYYYYYYYY&cccc=ppppppppppppppppppp&ddddd=ttttttttttttttttt&a=uuuuuuuuuuuuu",
	}
	for i, query := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			// Check against url.ParseQuery, ignoring the error.
			all, _ := url.ParseQuery(query)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for key := range all {
					_, _ = findFirstQueryKey(query, key)
				}
			}
		})
	}
}

func Benchmark_findQueryKeyGoLib(b *testing.B) {
	tests := []string{
		"a=1&b=2",
		"ascii=%3Ckey%3A+0x90%3E",
		"a=20&%20%3F&=%23+%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09:%2F@$%27%28%29%2A%2C%3B&a=30",
		"a=xxxxxxxxxxxxxxxx&bbb=YYYYYYYYYYYYYYY&cccc=ppppppppppppppppppp&ddddd=ttttttttttttttttt&a=uuuuuuuuuuuuu",
	}
	for i, query := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			// Check against url.ParseQuery, ignoring the error.
			all, _ := url.ParseQuery(query)
			var u url.URL
			u.RawQuery = query
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for key := range all {
					v := u.Query()[key]
					if len(v) > 0 {
						_ = v[0]
					}
				}
			}
		})
	}
}
