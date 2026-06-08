package thinkgo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Context represents the context of a single HTTP request.
// It wraps http.ResponseWriter and *http.Request, providing
// convenient methods for reading input and writing output.
//
// ThinkGo's Context is similar to ThinkPHP's controller context
// but more Go-idiomatic.
type Context struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	Params   map[string]string // URL path parameters
	Status   int
	bodyRead bool
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Writer:   w,
		Params:   make(map[string]string),
		Status:   http.StatusOK,
		bodyRead: false,
	}
}

func (c *Context) Get(key string) any {
	return c.Request.Context().Value(key)
}

func (c *Context) Set(key string, val any) {
	ctx := context.WithValue(c.Request.Context(), key, val)
	c.Request = c.Request.WithContext(ctx)
}

func (c *Context) JSON(data any) error {
	return NewResponse(c).JSON(data)
}

func (c *Context) Success(msg string, data ...any) error {
	return NewResponse(c).Success(msg, data...)
}

func (c *Context) Error(msg string, data ...any) error {
	return NewResponse(c).Error(msg, data...)
}

func (c *Context) Text(text string) error {
	return NewResponse(c).Text(text)
}

func (c *Context) HTML(html string) error {
	return NewResponse(c).HTML(html)
}

// Param returns a URL path parameter by name.
// Returns empty string if not found.
func (c *Context) Param(name string) string {
	if c.Params != nil {
		return c.Params[name]
	}
	return ""
}

// Query returns a query string parameter.
// Returns empty string if not found.
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}

// DefaultQuery returns a query parameter or a default value.
func (c *Context) DefaultQuery(name, defaultValue string) string {
	if v := c.Request.URL.Query().Get(name); v != "" {
		return v
	}
	return defaultValue
}

// QueryInt returns a query parameter as an int.
// Returns 0 if not found or not a valid integer.
func (c *Context) QueryInt(name string) int {
	val := c.Request.URL.Query().Get(name)
	if val == "" {
		return 0
	}
	var i int
	if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
		return 0
	}
	return i
}

// Form returns a form value by name.
func (c *Context) Form(name string) string {
	return c.Request.FormValue(name)
}

// BindJSON reads the request body and binds it to the given struct.
func (c *Context) BindJSON(v any) error {
	defer func() { c.bodyRead = true }()
	return json.NewDecoder(c.Request.Body).Decode(v)
}

// Header returns a request header value.
func (c *Context) Header(name string) string {
	return c.Request.Header.Get(name)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(name, value string) {
	c.Writer.Header().Set(name, value)
}

// SetStatus sets the HTTP response status code.
func (c *Context) SetStatus(code int) {
	c.Status = code
}

// Method returns the HTTP method.
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the request URL path.
func (c *Context) Path() string {
	return c.Request.URL.Path
}
