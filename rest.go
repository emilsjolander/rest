package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

	allSubmatches := varMatcher.FindAllStringSubmatch(pattern, -1)
	vars := make([]string, len(allSubmatches))
	for i, submatch := range allSubmatches {
		vars[i] = submatch[1]
	}

	m = &matcher{
		vars: vars,
		pattern: regexp.MustCompile(
			fmt.Sprintf(
				"^%s(/[A-Za-z0-9\\-_]*)*$",
				varMatcher.ReplaceAllString(pattern, "([A-Za-z0-9\\-_]+)"),
			),
		),
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
		default:
			ErrorHandler(w, &HttpError{Status: http.StatusInternalServerError, Message: "Internal server error"})
		}
	}
}

type SubRouter interface {
	http.Handler
	SetProcessedPath(string, *http.Request)
	ProcessedPath(*http.Request) string
}

type Routes map[string]http.Handler

func (this *Routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer RecoverError(w)

	prefix := this.ProcessedPath(r)
	path := r.URL.Path[len(prefix):]
	if path == "" {
		path = "/"
	}

	for key, value := range *this {
		matcher := matcherForPattern(key)
		if matches := matcher.pattern.FindStringSubmatch(path); matches != nil {
			// gather the arguments if any for this key
			matchString := key
			for i, v := range matcher.vars {
				match := matches[i+1] // skip first match as is is whole string
				matchString = strings.Replace(matchString, "{"+v+"}", match, -1)
				switch r.URL.RawQuery {
				case "":
					r.URL.RawQuery = fmt.Sprintf("%s=%s", v, match)
				default:
					r.URL.RawQuery += fmt.Sprintf("&%s=%s", v, match)
				}
			}

			// set the processed part of the path if the mapped handler is a SubRouter
			switch t := value.(type) {
			case SubRouter:
				t.SetProcessedPath(prefix+matchString, r)
				value.ServeHTTP(w, r)
				return
			default:
				// has to be exakt match, allowing trailing slash
				if strings.HasSuffix(path, matchString) || (path[len(path)-1] == '/' && strings.HasSuffix(path[:len(path)-1], matchString)) {
					value.ServeHTTP(w, r)
					return
				}
			}
		}
	}

	ErrorHandler(w, &HttpError{Status: http.StatusNotFound, Message: "No matching handler for this route"})
}

func (this *Routes) SetProcessedPath(prefix string, r *http.Request) {
	r.Header.Set("X-REST_PROCESSED_PATH", prefix)
}

func (this *Routes) ProcessedPath(r *http.Request) string {
	return r.Header.Get("X-REST_PROCESSED_PATH")
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
	switch strings.ToUpper(r.Method) {
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
