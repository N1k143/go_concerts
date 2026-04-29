package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"

	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	brightRed     = "\033[91m"
	brightGreen   = "\033[92m"
	brightYellow  = "\033[93m"
	brightBlue    = "\033[94m"
	brightMagenta = "\033[95m"
	brightCyan    = "\033[96m"
	brightWhite   = "\033[97m"

	bgRed  = "\033[41m"
	bgGreen = "\033[42m"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Status() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

func colorMethod(method string) string {
	var color string
	switch method {
	case http.MethodGet:
		color = brightGreen
	case http.MethodPost:
		color = brightBlue
	case http.MethodPut:
		color = brightYellow
	case http.MethodPatch:
		color = brightMagenta
	case http.MethodDelete:
		color = brightRed
	default:
		color = brightCyan
	}
	return fmt.Sprintf("%s%s%-7s%s", bold, color, method, reset)
}

func colorStatus(status int) string {
	var color string
	switch {
	case status < 200:
		color = brightCyan
	case status < 300:
		color = brightGreen
	case status < 400:
		color = brightYellow
	case status < 500:
		color = brightRed
	default:
		color = bgRed + white
	}
	return fmt.Sprintf("%s%s%d%s", bold, color, status, reset)
}

func colorLatency(d time.Duration) string {
	var color string
	switch {
	case d < 50*time.Millisecond:
		color = brightGreen
	case d < 200*time.Millisecond:
		color = brightYellow
	case d < 1*time.Second:
		color = yellow
	default:
		color = brightRed
	}
	return fmt.Sprintf("%s%s%s%s", bold, color, d.Round(time.Microsecond), reset)
}

func colorPath(path string) string {
	return fmt.Sprintf("%s%s%s", brightCyan, path, reset)
}

func colorTimestamp(t time.Time) string {
	return fmt.Sprintf("%s%s%s", magenta, t.Format("15:04:05.000"), reset)
}

// ColorLogger is a beautiful, colorful HTTP request logger middleware.
func ColorLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		status := wrapped.Status()

		separator := fmt.Sprintf("%s│%s", brightYellow, reset)

		fmt.Printf(
			"%s %s %s %s %s %s %s %s %s\n",
			colorTimestamp(start),
			separator,
			colorMethod(r.Method),
			separator,
			colorPath(r.URL.Path),
			separator,
			colorStatus(status),
			separator,
			colorLatency(duration),
		)
	})
}
