package thinkgo

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// Middleware is a function that wraps an http.Handler.
// It's the standard middleware signature in ThinkGo.
type Middleware func(next http.Handler) http.Handler

// MiddlewareFunc is an alias for convenience.
type MiddlewareFunc = Middleware

// MiddlewareGroup chains multiple middlewares into one.
func MiddlewareGroup(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Chain applies middlewares in order and returns the final handler.
// First middleware in the list is the outermost.
func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// --- Built-in Middleware ---

// Recovery returns a middleware that recovers from panics and logs the stack trace.
// Returns 500 JSON response with error message.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log stack trace
					slog.Error("panic recovered",
						"method", r.Method,
						"path", r.URL.Path,
						"error", err,
						"stack", string(debug.Stack()),
					)

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"code":500,"msg":"Internal Server Error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// LoggerMW returns a middleware that logs requests with duration.
func LoggerMW() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)

			slog.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
				"duration", duration.String(),
			)
		})
	}
}

// CORSMiddleware returns a CORS middleware.
// Allows all origins by default (customize via WithCORS).
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithCORS returns a CORS middleware with custom configuration.
//
//	router.Use(thinkgo.WithCORS("https://myapp.com", "GET,POST", "Authorization"))
func WithCORS(origin, methods, headers string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestID returns a middleware that adds a unique request ID to each request.
// If the client sends X-Request-Id, it's passed through; otherwise a new one is generated.
//
// Access the request ID via:
//
//	ctx.Get("request_id")
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = Token(16) // uses crypto/rand, generates 32-char hex
			}

			// Add to response header
			w.Header().Set("X-Request-Id", id)

			// Store in request context
			ctx := context.WithValue(r.Context(), "request_id", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Timeout returns a middleware that aborts requests exceeding the given duration.
// The handler receives a context with deadline.
//
//	router.Use(thinkgo.Timeout(30 * time.Second))
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write([]byte(`{"code":504,"msg":"Request timeout"}`))
			}
		})
	}
}

// BasicAuth returns a middleware that validates HTTP Basic Authentication.
//
//	router.Use(thinkgo.BasicAuth("admin", "secret123"))
func BasicAuth(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || u != username || p != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":401,"msg":"Unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
