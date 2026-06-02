package middleware

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// RequestLogger records high-level metadata about each HTTP request without logging bodies.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w}

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			logger.LogAttrs(r.Context(), requestLogLevel(lrw.Status()), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", lrw.Status()),
				slog.Duration("duration", duration),
				slog.Int64("bytes", lrw.BytesWritten()),
				slog.String("client_ip", clientIP(r)),
				slog.String("forwarded_for", forwardedFor(r)),
			)
		})
	}
}

func requestLogLevel(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func forwardedFor(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if header == "" {
		return ""
	}

	parts := strings.Split(header, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return strings.Join(parts, ",")
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
}

func (lrw *loggingResponseWriter) Status() int {
	if lrw.status == 0 {
		return http.StatusOK
	}
	return lrw.status
}

func (lrw *loggingResponseWriter) BytesWritten() int64 {
	return lrw.bytesWritten
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	if lrw.status == 0 {
		lrw.status = statusCode
	}
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytesWritten += int64(n)
	return n, err
}

func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := lrw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (lrw *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := lrw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
