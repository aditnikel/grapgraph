package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/aditnikel/grapgraph/src/observability"
)

func NewRecoverer(l *observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					l.Error("panic_recovered", observability.Fields{
						"panic": rec,
						"stack": string(debug.Stack()),
						"path":  r.URL.Path,
					})
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
