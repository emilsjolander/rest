rest
====
A very plain and simple web request router for go. The goal was to make something that was both fun, easy, and flexible to use while respecting the great type system of the go language. Everything is built around the `http.Handler` interface which makes it very easy to write all kinds of middleware. One this that might not be very idomatic but i have found really cleans my code up is to report errors via panics. This allows me to right functions such as `RequireUser()` which will act like before filters in other libraries.

example
-------
```go

import (
	"fmt"
	"net/http"

	"github.com/emilsjolander/rest"
)

func main() {
	routes := &rest.Routes{
		"/": &rest.Methods{
			Get: http.HandlerFunc(Welcome),
		},
		"/hello": &rest.Routes{
			"/": &rest.Methods{
				Get: http.HandlerFunc(Hello),
			},
			"/{name}": &rest.Methods{
				Get: http.HandlerFunc(HelloThere),
			},
		},
	}

	http.Handle("/", routes)
	http.ListenAndServe(":8080", nil)
}

func Welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func HelloThere(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello "+r.URL.Query().Get("name"))
}

```
