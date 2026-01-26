package middleware

import (
	"net/http"
	"time"

	"github.com/aditnikel/grapgraph/src/observability"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(p []byte) (int, error) {
	if s.status == 0 {
		s.status = 200
	}
	n, err := s.ResponseWriter.Write(p)
	s.bytes += n
	return n, err
}

func NewRequestLog(l *observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(sw, r)

			l.Info("http_request", observability.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": sw.status,
				"bytes":  sw.bytes,
				"dur_ms": time.Since(start).Milliseconds(),
			})
		})
	}
}
