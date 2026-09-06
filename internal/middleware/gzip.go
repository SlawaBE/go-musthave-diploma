package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/SlawaBE/go-musthave-diploma/internal/gzip"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rr *responseRecorder) Header() http.Header {
	return rr.ResponseWriter.Header()
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	return rr.body.Write(b)
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.status = statusCode
}

func GZip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()
			r.Body = io.NopCloser(gzipReader)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}

		recorder := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		handler.ServeHTTP(recorder, r)

		contentType := recorder.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			gzipWriter := gzip.NewCompressWriter(w)
			defer gzipWriter.Close()

			gzipWriter.WriteHeader(recorder.status)

			_, err := gzipWriter.Write(recorder.body.Bytes())
			if err != nil {
				http.Error(w, "failed to compres response", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(recorder.status)
		_, err := w.Write(recorder.body.Bytes())
		if err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
		}
	})
}
