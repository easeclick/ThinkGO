package thinkgo

import (
	"fmt"
	"net/http"
	"sort"
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
//
// Supports:
//   - Path parameters (:id, :slug)
//   - Route groups with shared prefix + middleware
//   - Nested groups at any depth
//   - RESTful resource routing
//   - 405 Method Not Allowed
type Router struct {
	routes     []RouteInfo
	subRouters []*Router
	prefix     string
	middleware []Middleware
	notFound   HandlerFunc
	methodNotAllowed HandlerFunc
}

// NewRouter creates a new router.
func NewRouter() *Router {
	r := &Router{
		routes:     make([]RouteInfo, 0),
		subRouters: make([]*Router, 0),
		middleware: make([]Middleware, 0),
	}
	r.notFound = func(c *Context) error {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(c.Writer, `{"code":404,"msg":"Route not found","request":"%s"}`, c.Path())
		return nil
	}
	r.methodNotAllowed = func(c *Context) error {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintf(c.Writer, `{"code":405,"msg":"Method not allowed","request":"%s %s"}`, c.Method(), c.Path())
		return nil
	}
	return r
}

// Group creates a route group with a prefix and optional middleware.
// Groups can be nested at any depth.
//
//	api := r.Group("/api", thinkgo.AuthMiddleware())
//	api.Get("/users", listUsers)  // → GET /api/users
func (r *Router) Group(prefix string, mws ...Middleware) *Router {
	sub := NewRouter()
	sub.prefix = r.prefix + prefix
	sub.middleware = append(append([]Middleware{}, r.middleware...), mws...)
	sub.notFound = r.notFound
	sub.methodNotAllowed = r.methodNotAllowed
	r.subRouters = append(r.subRouters, sub)
	return sub
}

// Use adds middleware to the router (applied to all routes in this group/root).
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

// Patch registers a PATCH route.
func (r *Router) Patch(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("PATCH", pattern, handler, mws...)
}

// Options registers an OPTIONS route.
func (r *Router) Options(pattern string, handler HandlerFunc, mws ...Middleware) {
	r.addRoute("OPTIONS", pattern, handler, mws...)
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
//	GET    /resource      → Index
//	GET    /resource/:id  → Show
//	POST   /resource      → Store
//	PUT    /resource/:id  → Update
//	DELETE /resource/:id  → Delete
func (r *Router) Resource(name string, handler any) {
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

// MatchResult holds the result of route matching.
type MatchResult struct {
	Route  RouteInfo
	Params map[string]string
}

// resolveRoute recursively searches the router tree for a matching route.
// Returns the matched route + params, or nil + list of allowed methods for 405.
func (r *Router) resolveRoute(method, path string) (*MatchResult, map[string]bool) {
	var allowedMethods map[string]bool

	// Search current router's route table
	for _, route := range r.routes {
		params, ok := matchPath(route.Pattern, path)
		if !ok {
			continue
		}
		// Pattern matched — check method
		if route.Method == method {
			return &MatchResult{Route: route, Params: params}, nil
		}
		// Wrong method, collect allowed methods
		if allowedMethods == nil {
			allowedMethods = make(map[string]bool)
		}
		allowedMethods[route.Method] = true
	}

	// Recursively search sub-routers (depth-first)
	for _, sub := range r.subRouters {
		match, subAllowed := sub.resolveRoute(method, path)
		if match != nil {
			return match, nil
		}
		for m := range subAllowed {
			if allowedMethods == nil {
				allowedMethods = make(map[string]bool)
			}
			allowedMethods[m] = true
		}
	}

	return nil, allowedMethods
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimRight(req.URL.Path, "/")
	if path == "" {
		path = "/"
	}

	ctx := NewContext(w, req)
	match, allowed := r.resolveRoute(req.Method, path)

	if match != nil {
		ctx.Params = match.Params
		r.serveWithMiddleware(w, req, ctx, match.Route)
		return
	}

	// 405 Method Not Allowed
	if len(allowed) > 0 {
		methods := make([]string, 0, len(allowed))
		for m := range allowed {
			methods = append(methods, m)
		}
		sort.Strings(methods)
		w.Header().Set("Allow", strings.Join(methods, ", "))
		_ = r.methodNotAllowed(ctx)
		return
	}

	// 404 Not Found
	_ = r.notFound(ctx)
}

// serveWithMiddleware runs the handler through its middleware chain.
func (r *Router) serveWithMiddleware(w http.ResponseWriter, req *http.Request, ctx *Context, route RouteInfo) {
	var final http.Handler
	final = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.Request = r // pick up modified request from middleware chain
		if err := route.Handler(ctx); err != nil {
			// Handler returned an error — write as JSON 500 unless already handled
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"code":500,"msg":"%s"}`, err.Error())
		}
	})

	for i := len(route.Middleware) - 1; i >= 0; i-- {
		final = route.Middleware[i](final)
	}

	final.ServeHTTP(w, req)
}

// PrintRoutes prints all registered routes for debugging.
func (r *Router) PrintRoutes() {
	fmt.Println("\n┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Registered Routes                                         │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	// Collect all routes recursively
	var all []RouteInfo
	r.collectRoutes(&all)

	// Sort by pattern then method
	sort.Slice(all, func(i, j int) bool {
		if all[i].Pattern != all[j].Pattern {
			return all[i].Pattern < all[j].Pattern
		}
		return all[i].Method < all[j].Method
	})

	for _, route := range all {
		fmt.Printf("│  %-6s %-45s │\n", route.Method, route.Pattern)
	}
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Printf("  Total: %d route(s)\n\n", len(all))
}

// collectRoutes recursively collects all routes from the router tree.
func (r *Router) collectRoutes(routes *[]RouteInfo) {
	*routes = append(*routes, r.routes...)
	for _, sub := range r.subRouters {
		sub.collectRoutes(routes)
	}
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

	return params, true
}

// StdHandler converts a HandlerFunc to http.HandlerFunc.
func StdHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		_ = fn(ctx)
	}
}
