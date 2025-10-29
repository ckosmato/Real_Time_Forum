package middleware

import (
	"fmt"
	"net/http"
)

type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// Log wraps an http.Handler and returns a new http.Handler that logs the request
// before delegating to the next handler
func (m *LoggingMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.Proto)
		next.ServeHTTP(w, r)
	})
}