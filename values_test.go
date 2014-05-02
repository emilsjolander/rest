package rest

import (
	"net/http"
	"testing"
)

func TestRequire(t *testing.T) {
	var name string
	routes := &Routes{
		"/hi": &Routes{
			"/there/{first_name}/{last_name}": &Methods{
				Get: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					name = Params(r).Require("first_name").String() + Params(r).Require("last_name").String()
				}),
			},
		},
	}

	r, _ := http.NewRequest("GET", "htttp://www.test.com/hi/there/bob/bobster", nil)
	routes.ServeHTTP(nil, r)
	if name != "bobbobster" {
		t.Fail()
	}
}
