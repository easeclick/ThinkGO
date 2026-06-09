package thinkgo

import (
	"encoding/json"
	"net/http"
)

// Response provides helper methods for writing HTTP responses.
// Similar to ThinkPHP's response handling but Go-native.
type Response struct {
	ctx *Context
}

// NewResponse creates a Response helper.
func NewResponse(ctx *Context) *Response {
	return &Response{ctx: ctx}
}

// JSON writes a JSON response using ctx.Status as the status code.
func (r *Response) JSON(data any) error {
	r.ctx.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.ctx.Writer.WriteHeader(r.ctx.Status)
	return json.NewEncoder(r.ctx.Writer).Encode(data)
}

// CodeJSON writes a JSON response with an explicit HTTP status code.
//
//	response.CodeJSON(http.StatusOK, data)
func (r *Response) CodeJSON(code int, data any) error {
	r.ctx.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.ctx.Writer.WriteHeader(code)
	return json.NewEncoder(r.ctx.Writer).Encode(data)
}

// JSONP writes a JSONP response.
func (r *Response) JSONP(callback string, data any) error {
	r.ctx.Writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	r.ctx.Writer.WriteHeader(r.ctx.Status)

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.ctx.Writer.Write([]byte(callback + "(" + string(bs) + ")"))
	return err
}

// XML writes an XML response (placeholder for now).
func (r *Response) XML(data any) error {
	r.ctx.Writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	r.ctx.Writer.WriteHeader(r.ctx.Status)
	return json.NewEncoder(r.ctx.Writer).Encode(data)
}

// Text writes a plain text response.
func (r *Response) Text(text string) error {
	r.ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.ctx.Writer.WriteHeader(r.ctx.Status)
	_, err := r.ctx.Writer.Write([]byte(text))
	return err
}

// HTML writes an HTML response.
func (r *Response) HTML(html string) error {
	r.ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ctx.Writer.WriteHeader(r.ctx.Status)
	_, err := r.ctx.Writer.Write([]byte(html))
	return err
}

// NoContent sends a 204 No Content response.
func (r *Response) NoContent() error {
	r.ctx.Writer.WriteHeader(http.StatusNoContent)
	return nil
}

// Redirect sends a redirect response.
func (r *Response) Redirect(url string, code ...int) {
	status := http.StatusFound
	if len(code) > 0 {
		status = code[0]
	}
	http.Redirect(r.ctx.Writer, r.ctx.Request, url, status)
}

// --- ThinkPHP-style response helpers ---

// Success sends a ThinkPHP-style success response (code=1).
//
//	curl response: {"code":1, "msg":"success", "data":...}
func (r *Response) Success(msg string, data ...any) error {
	resp := map[string]any{
		"code": 1,
		"msg":  msg,
	}
	if len(data) > 0 {
		resp["data"] = data[0]
	}
	return r.JSON(resp)
}

// Error sends a ThinkPHP-style error response (code=0).
//
//	curl response: {"code":0, "msg":"error message"}
func (r *Response) Error(msg string, data ...any) error {
	resp := map[string]any{
		"code": 0,
		"msg":  msg,
	}
	if len(data) > 0 {
		resp["data"] = data[0]
	}
	// Default to 400 status for errors (caller can change via ctx.SetStatus first)
	if r.ctx.Status == http.StatusOK {
		r.ctx.Status = http.StatusBadRequest
	}
	return r.JSON(resp)
}

// Fail is an alias for Error (more semantic: "request failed").
func (r *Response) Fail(msg string, data ...any) error {
	return r.Error(msg, data...)
}
