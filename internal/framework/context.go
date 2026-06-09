package thinkgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
	keys     map[string]any
	keysMu   sync.RWMutex
	bodyRead bool
	aborted  bool
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Writer:   w,
		Params:   make(map[string]string),
		Status:   http.StatusOK,
		keys:     make(map[string]any),
		bodyRead: false,
		aborted:  false,
	}
}

// --- Key-value store (per-request, replaces context.WithValue abuse) ---

// Get retrieves a value set by Set.
// Returns nil if not found.
//
// Unlike ThinkPHP's session/request get, this is a per-request key-value store.
func (c *Context) Get(key string) any {
	c.keysMu.RLock()
	defer c.keysMu.RUnlock()
	return c.keys[key]
}

// Set stores a value in the per-request context.
// Values are visible for the duration of the request only.
func (c *Context) Set(key string, val any) {
	c.keysMu.Lock()
	defer c.keysMu.Unlock()
	c.keys[key] = val
}

// Has checks if a key exists in the context.
func (c *Context) Has(key string) bool {
	c.keysMu.RLock()
	defer c.keysMu.RUnlock()
	_, ok := c.keys[key]
	return ok
}

// --- Abort mechanism (stops middleware chain) ---

// Abort marks the context as aborted, preventing further middleware from executing.
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted returns whether the request has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}

// --- HTTP method checks (like ThinkPHP's Request::isGet/isPost) ---

// IsGet returns true if the request method is GET.
func (c *Context) IsGet() bool {
	return c.Method() == http.MethodGet
}

// IsPost returns true if the request method is POST.
func (c *Context) IsPost() bool {
	return c.Method() == http.MethodPost
}

// IsPut returns true if the request method is PUT.
func (c *Context) IsPut() bool {
	return c.Method() == http.MethodPut
}

// IsDelete returns true if the request method is DELETE.
func (c *Context) IsDelete() bool {
	return c.Method() == http.MethodDelete
}

// IsPatch returns true if the request method is PATCH.
func (c *Context) IsPatch() bool {
	return c.Method() == http.MethodPatch
}

// IsOptions returns true if the request method is OPTIONS.
func (c *Context) IsOptions() bool {
	return c.Method() == http.MethodOptions
}

// IsAjax returns true if the request has X-Requested-With: XMLHttpRequest header.
func (c *Context) IsAjax() bool {
	return strings.EqualFold(c.Header("X-Requested-With"), "XMLHttpRequest")
}

// --- Response helpers (delegate to Response) ---

// JSON sends a JSON response with current status code.
func (c *Context) JSON(data any) error {
	return NewResponse(c).JSON(data)
}

// Success sends a ThinkPHP-style success response.
func (c *Context) Success(msg string, data ...any) error {
	return NewResponse(c).Success(msg, data...)
}

// Error sends a ThinkPHP-style error response.
func (c *Context) Error(msg string, data ...any) error {
	return NewResponse(c).Error(msg, data...)
}

// Text writes a plain text response.
func (c *Context) Text(text string) error {
	return NewResponse(c).Text(text)
}

// HTML writes an HTML response.
func (c *Context) HTML(html string) error {
	return NewResponse(c).HTML(html)
}

// Redirect sends a redirect response.
func (c *Context) Redirect(url string, code ...int) {
	NewResponse(c).Redirect(url, code...)
}

// Write writes raw bytes to the response.
func (c *Context) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

// --- Input helpers ---

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

// QuerySlice returns multiple values for a query parameter (e.g. ?id=1&id=2).
func (c *Context) QuerySlice(name string) []string {
	vals, ok := c.Request.URL.Query()[name]
	if !ok {
		return nil
	}
	return vals
}

// Form returns a form value by name.
func (c *Context) Form(name string) string {
	return c.Request.FormValue(name)
}

// FormSlice returns multiple values for a form parameter.
func (c *Context) FormSlice(name string) []string {
	c.Request.ParseMultipartForm(32 << 20) // 32MB
	vals, ok := c.Request.Form[name]
	if !ok {
		return nil
	}
	return vals
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

// RemoteIP returns the client's IP address.
func (c *Context) RemoteIP() string {
	// Check X-Forwarded-For first
	if fwd := c.Header("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	// Check X-Real-IP
	if realIP := c.Header("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fall back to RemoteAddr (strip port)
	addr := c.Request.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		return addr[:idx]
	}
	return addr
}

// UserAgent returns the User-Agent header.
func (c *Context) UserAgent() string {
	return c.Header("User-Agent")
}

// ContentType returns the Content-Type header.
func (c *Context) ContentType() string {
	return c.Header("Content-Type")
}
