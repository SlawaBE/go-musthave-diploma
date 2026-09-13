package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func RequestLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		handler.ServeHTTP(rw, r)

		logger.Log.Info("receive http request",
			slog.String("uri", r.RequestURI),
			slog.String("method", r.Method),
			slog.String("duration", time.Since(start).String()),
			slog.Int("status", rw.status),
			slog.Int("size", rw.size),
		)
	})
}
