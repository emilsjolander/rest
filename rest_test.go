package rest

import (
	"net/http"
	"testing"
)

func TestRoute(t *testing.T) {
	matched := ""
	routes := &Routes{
		"/": &Methods{
			Get: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				matched = "/"
			}),
		},
		"/hi": &Routes{
			"/": &Methods{
				Get: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					matched = "/hi"
				}),
			},
			"/there/{first_name}/{last_name}": &Methods{
				Get: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					matched = "/hi/there"
				}),
			},
		},
	}

	ErrorHandler = func(w http.ResponseWriter, err *HttpError) {
		matched = "error"
	}

	var r *http.Request

	r, _ = http.NewRequest("GET", "htttp://www.test.com/", nil)
	routes.ServeHTTP(nil, r)
	if matched != "/" {
		t.Error("root route did not match, matched = " + matched)
	}

	r, _ = http.NewRequest("GET", "htttp://www.test.com/hi/", nil)
	routes.ServeHTTP(nil, r)
	if matched != "/hi" {
		t.Error("/hi route did not match, matched = " + matched)
	}

	r, _ = http.NewRequest("GET", "htttp://www.test.com/hi/there/bob/bobster", nil)
	routes.ServeHTTP(nil, r)
	if matched != "/hi/there" {
		t.Error("/hi/there route did not match, matched = " + matched)
	}

	r, _ = http.NewRequest("GET", "htttp://www.test.com/hi/there/error", nil)
	routes.ServeHTTP(nil, r)
	if matched != "error" {
		t.Error("error handler did not match, matched = " + matched)
	}
}
