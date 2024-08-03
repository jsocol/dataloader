package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

type wrappedWriter struct {
	http.ResponseWriter
	mu      sync.Mutex
	status  int
	written int
}

func (w *wrappedWriter) WriteHeader(status int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *wrappedWriter) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err = w.ResponseWriter.Write(b)
	w.written = n
	return
}

type Logger struct {
	target http.Handler
}

func WithLogger(h http.Handler) http.Handler {
	return &Logger{
		target: h,
	}
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ww := &wrappedWriter{ResponseWriter: w}
	defer func() {
		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.Int("status", ww.status),
			slog.String("path", r.URL.Path),
			slog.Int("response.bytes", ww.written),
			slog.String("client.ip", r.RemoteAddr),
		}
		level := slog.LevelInfo
		if ww.status < 100 || ww.status >= 500 {
			level = slog.LevelError
		}
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		msg := fmt.Sprintf("%s - %s %s [%d] %db", r.RemoteAddr, r.Method, r.URL.Path+query, ww.status, ww.written)
		slog.LogAttrs(r.Context(), level, msg, attrs...)
	}()
	l.target.ServeHTTP(ww, r)
}
