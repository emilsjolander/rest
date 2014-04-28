package rest

import (
	"net/http"
	"net/url"
	"strconv"
)

type Value string

func (v Value) String() string {
	return string(v)
}

func (v Value) Int() int {
	return int(v.Int64())
}

func (v Value) Int64() int64 {
	i, err := strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		panic(&HttpError{Status: 400, Message: "Invalid argument type: " + string(v)})
	}
	return i
}

func (v Value) Float() float64 {
	f, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		panic(&HttpError{Status: 400, Message: "Invalid argument type: " + string(v)})
	}
	return f
}

func (v Value) Bool() bool {
	b, err := strconv.ParseBool(string(v))
	if err != nil {
		panic(&HttpError{Status: 400, Message: "Invalid argument type: " + string(v)})
	}
	return b
}

type Values struct {
	values url.Values
}

func (v *Values) Require(key string) Value {
	_, ok := v.values[key]
	if !ok {
		panic(&HttpError{Status: 400, Message: "Missing argument: " + key})
	}
	return Value(v.values.Get(key))
}

func (v *Values) Optional(key string, def string) Value {
	_, ok := v.values[key]
	if !ok {
		return Value(def)
	}
	return Value(v.values.Get(key))
}

func Query(r *http.Request) *Values {
	return &Values{r.URL.Query()}
}

func Form(r *http.Request) *Values {
	r.ParseForm()
	return &Values{r.Form}
}

func Params(r *http.Request) *Values {
	r.ParseForm()
	values := &Values{r.Form}
	for key, value := range r.URL.Query() {
		values.values.Add(key, value[0])
	}
	return values
}
