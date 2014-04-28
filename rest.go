package rest

import (
	"fmt"
	"net/http"
	"regexp"
)

type matcher struct {
	vars    []string
	pattern *regexp.Regexp
}

var varMatcher = regexp.MustCompile("\\{([A-Za-z0-9\\-_]+)\\}")
var matcherCache = make(map[string]*matcher)

func matcherForPattern(pattern string) *matcher {
	m, ok := matcherCache[pattern]
	if ok {
		return m
	}

	m = &matcher{
		vars:    varMatcher.FindStringSubmatch(pattern),
		pattern: regexp.MustCompile("^" + varMatcher.ReplaceAllString(pattern, "([A-Za-z0-9\\-_]+)") + "(/[A-Za-z0-9\\-_]*)*$"),
	}

	matcherCache[pattern] = m
	return m
}

type HttpError struct {
	Status  int
	Message string
}

func (this *HttpError) Error() string {
	return fmt.Sprintf("%d: %s", this.Status, this.Message)
}

var ErrorHandler = func(w http.ResponseWriter, err *HttpError) {
	w.WriteHeader(err.Status)
	fmt.Fprintf(w, "%v", err)
}

func RecoverError(w http.ResponseWriter) {
	if err := recover(); err != nil {
		switch t := err.(type) {
		case *HttpError:
			ErrorHandler(w, t)
		case error:
			ErrorHandler(w, &HttpError{Status: http.StatusInternalServerError, Message: t.Error()})
		case string:
			ErrorHandler(w, &HttpError{Status: http.StatusInternalServerError, Message: t})
		default:
			ErrorHandler(w, &HttpError{Status: http.StatusInternalServerError, Message: "Internal server error"})
		}
	}
}

type Prefixable interface {
	http.Handler
	SetPrefix(string, *http.Request)
	GetPrefix(*http.Request) string
}

type PrefixHandler struct {
	Prefix  string
	Handler Prefixable
}

func (this *PrefixHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer RecoverError(w)
	this.Handler.SetPrefix(this.Prefix, r)
	this.Handler.ServeHTTP(w, r)
}

type Routes map[string]http.Handler

func (this *Routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer RecoverError(w)

	prefix := this.GetPrefix(r)
	path := r.URL.Path[len(prefix):]
	for key, value := range *this {
		matcher := matcherForPattern(key)
		if matches := matcher.pattern.FindStringSubmatch(path); matches != nil {
			for i, v := range matcher.vars {
				switch r.URL.RawQuery {
				case "":
					r.URL.RawQuery = fmt.Sprintf("?%s=%s", v, matches[i])
				default:
					r.URL.RawQuery += fmt.Sprintf("&%s=%s", v, matches[i])
				}
			}
			switch t := value.(type) {
			case Prefixable:
				t.SetPrefix(prefix+key, r)
			}
			value.ServeHTTP(w, r)
			return
		}
	}
	ErrorHandler(w, &HttpError{Status: http.StatusNotFound, Message: "No matching handler for this route"})
}

func (this *Routes) SetPrefix(prefix string, r *http.Request) {
	r.Header.Set("__REST_PREFIX_HEADER__", prefix)
}

func (this *Routes) GetPrefix(r *http.Request) string {
	return r.Header.Get("__REST_PREFIX_HEADER__")
}

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

func (this *Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer RecoverError(w)

	var handler http.Handler
	switch r.Method {
	case "GET":
		handler = this.Get
	case "POST":
		handler = this.Post
	case "PUT":
		handler = this.Put
	case "DELETE":
		handler = this.Delete
	case "PATCH":
		handler = this.Patch
	case "HEAD":
		handler = this.Head
	case "OPTIONS":
		handler = this.Options
	case "CONNECT":
		handler = this.Connect
	case "TRACE":
		handler = this.Trace
	}
	if handler == nil {
		ErrorHandler(w, &HttpError{Status: http.StatusNotFound, Message: "No matching handler for this route"})
		return
	}
	handler.ServeHTTP(w, r)
}
