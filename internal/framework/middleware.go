package thinkgo

import (
	"log/slog"
	"net/http"
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

// Recovery returns a middleware that recovers from panics.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// LoggerMW returns a middleware that logs requests.
func LoggerMW() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware returns a CORS middleware.
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
