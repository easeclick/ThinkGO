package thinkgo

import (
	"fmt"
	"net/http"
	"strings"
)

// HandlerFunc is the main handler signature in ThinkGo.
// Similar to ThinkPHP's controller action.
type HandlerFunc func(*Context) error

// RouteInfo holds a registered route.
type RouteInfo struct {
	Method     string
	Pattern    string
	Handler    HandlerFunc
	Middleware []Middleware
}

// Router manages route registration and dispatch.
// Inspired by ThinkPHP's route system.
type Router struct {
	routes      []RouteInfo
	subRouters  []*Router
	prefix      string
	middleware  []Middleware
	notFound    HandlerFunc
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		routes:      make([]RouteInfo, 0),
		subRouters:  make([]*Router, 0),
		middleware:  make([]Middleware, 0),
		notFound: func(c *Context) error {
			c.SetStatus(http.StatusNotFound)
			return NewResponse(c).JSON(map[string]any{
				"code":    0,
				"msg":     "Route not found",
				"request": c.Path(),
			})
		},
	}
}

// Group creates a route group with a prefix and optional middleware.
func (r *Router) Group(prefix string, mws ...Middleware) *Router {
	sub := NewRouter()
	sub.prefix = r.prefix + prefix
	sub.middleware = append(append([]Middleware{}, r.middleware...), mws...)
	sub.notFound = r.notFound
	r.subRouters = append(r.subRouters, sub)
	return sub
}

// Use adds middleware to the router (applied to all routes).
func (r *Router) Use(mws ...Middleware) {
	r.middleware = append(r.middleware, mws...)
}

// Get registers a GET route.
func (r *Router) Get(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("GET", pattern, handler, mws...)
}

// Post registers a POST route.
func (r *Router) Post(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("POST", pattern, handler, mws...)
}

// Put registers a PUT route.
func (r *Router) Put(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("PUT", pattern, handler, mws...)
}

// Delete registers a DELETE route.
func (r *Router) Delete(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("DELETE", pattern, handler, mws...)
}

// Any registers a route that matches all HTTP methods.
func (r *Router) Any(pattern string, handler HandlerFunc, mws ...Middleware) {
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"} {
		r.addRoute(method, pattern, handler, mws...)
	}
}

// Resource registers RESTful resource routes.
// Like ThinkPHP's Route::resource():
//
//	GET    /resource      → index
//	GET    /resource/:id  → show
//	POST   /resource      → store
//	PUT    /resource/:id  → update
//	DELETE /resource/:id  → delete
func (r *Router) Resource(name string, handler any) {
	// handler should implement ResourceHandler interface
	// We use type assertion to call methods
	h, ok := handler.(ResourceHandler)
	if !ok {
		panic(fmt.Sprintf("route: %s does not implement ResourceHandler", name))
	}

	base := "/" + strings.Trim(name, "/")
	item := base + "/:id"

	r.Get(base, h.Index)
	r.Post(base, h.Store)
	r.Get(item, h.Show)
	r.Put(item, h.Update)
	r.Delete(item, h.Delete)
}

// ResourceHandler defines the methods for a RESTful resource controller.
type ResourceHandler interface {
	Index(*Context) error
	Store(*Context) error
	Show(*Context) error
	Update(*Context) error
	Delete(*Context) error
}

// addRoute is the internal route registration method.
func (r *Router) addRoute(method, pattern string, handler HandlerFunc, mws ...Middleware) {
	fullPattern := r.prefix + pattern

	// Merge router-level middleware with route-level middleware
	allMws := make([]Middleware, 0, len(r.middleware)+len(mws))
	allMws = append(allMws, r.middleware...)
	allMws = append(allMws, mws...)

	r.routes = append(r.routes, RouteInfo{
		Method:     method,
		Pattern:    fullPattern,
		Handler:    handler,
		Middleware: allMws,
	})
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := NewContext(w, req)
	path := strings.TrimRight(req.URL.Path, "/")
	if path == "" {
		path = "/"
	}

	matched := r.tryServe(w, req, ctx, path)
	if !matched {
		for _, sub := range r.subRouters {
			if sub.tryServe(w, req, ctx, path) {
				return
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"code":0,"msg":"Route not found","request":"%s"}`, path)
	}
}

// tryServe tries to match and serve a route from this router's route table.
// Returns true if a route was matched and served.
func (r *Router) tryServe(w http.ResponseWriter, req *http.Request, ctx *Context, path string) bool {
	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}

		params, ok := matchPath(route.Pattern, path)
		if !ok {
			continue
		}
		ctx.Params = params

		var final http.Handler
		final = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r // pick up modified request from middleware chain
			if err := route.Handler(ctx); err != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"code":0,"msg":"%s"}`, err.Error())
			}
		})

		for i := len(route.Middleware) - 1; i >= 0; i-- {
			final = route.Middleware[i](final)
		}

		final.ServeHTTP(w, req)
		return true
	}
	return false
}

// matchPath matches a pattern like "/user/:id" against a path like "/user/123".
// Returns extracted parameters and whether it matched.
func matchPath(pattern, path string) (map[string]string, bool) {
	// Normalize
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		pattern = "/"
	}

	// Handle exact match (no params)
	if !strings.Contains(pattern, ":") {
		return nil, pattern == path
	}

	parts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(parts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}

	return params, len(params) > 0
}

// StdHandler converts a HandlerFunc to http.HandlerFunc.
func StdHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		_ = fn(ctx)
	}
}
