rest
====
A simple web request router for go. The goal with rest is to enable flexible request routing while respecting the great type system of the Go language. Everything is built around the `http.Handler` interface which makes your code both type safe and easy to integrate with other libraries. There is no use of the `interface{}` type in the library as keeping Go's static typing is one of the main goals. One thing that might not be very idomatic but i have found really cleans my code up is to report errors in handlers via panics. This allows you to write functions such as `RequireUser(*http.Request)` which will act like a before filters.

Example
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
			"/{user_id}": &rest.Methods{
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
	fmt.Fprintf(w, "Hello " + rest.Params(r).Require("name").String())
}
```

Api
---
###rest.Routes
`rest.Routes` is a map type implementing the `http.Handler` interface. `rest.Routes` maps string (url patterns) to `http.Handler` types. because a `rest.Routes` is a `http.Handler` you can easily create sub-routes by using another `rest.Routes` as the `http.handler` being mapped to.

```go
type Routes map[string]http.Handler
```

###rest.Methods
`rest.Methods` is a struct type with handler properties for every http method. Attach any `http.Handler` to the different methods to route a specific method to a specific handler. As with every other type `rest.Methods` also implements the `http.Handler` interface.

```go
type Methods struct {
	Get     http.Handler
	Post    http.Handler
	Put     http.Handler
	Delete  http.Handler
	Patch   http.Handler
	Head    http.Handler
	Options http.Handler
	Connect http.Handler
	Trace   http.Handler
}
```

###rest.SubRouter
`rest.SubRouter` is an interface which defines a `http.Handler` that is not a leaf node in your request handler tree. To implement `rest.SubRouter` you must implement `http.Handler` as will as `SetProcessedPath(string, *http.Request)` and `ocessedPath(*http.Request) string`. `rest.Router` implements this interface.

```go
type SubRouter interface {
	http.Handler
	SetProcessedPath(string, *http.Request)
	ProcessedPath(*http.Request) string
}
```

###rest.Values
A `rest.Values` is a struct that wraps `url.Values` and can be obtained through `rest.Params(*http.Request)`, `rest.Query(*http.Request)`, or `rest.Form(*http.Request)`. A `rest.Values` instance lets you make sure that certain values exist is the query string, the form values or any of them (`rest.Params(*http.Request)`).

```go
func Query(r *http.Request) *Values
func Form(r *http.Request) *Values
func Params(r *http.Request) *Values

func (v *Values) Require(key string) Value
func (v *Values) Optional(key string, def string) Value
```