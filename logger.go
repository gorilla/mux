package mux

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type LogConfig struct {
	Output io.Writer
}

type LogFormatterParams struct {
	Request *http.Request

	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func (p *LogFormatterParams) MethodColor() string {
	method := p.Method

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

func Logger(next http.Handler) http.Handler {
	return LoggerWithConfig(LogConfig{})(next)
}

func LoggerWithConfig(c LogConfig) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		out := c.Output
		if out == nil {
			out = os.Stdout
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			sw := statusWriter{
				ResponseWriter: w,
				status:         0,
			}
			next.ServeHTTP(&sw, r)

			if raw != "" {
				path = fmt.Sprintf("%s?%s", path, raw)
			}

			stop := time.Now()
			p := LogFormatterParams{
				Request:    r,
				TimeStamp:  stop,
				Latency:    stop.Sub(start),
				Method:     r.Method,
				StatusCode: sw.status,
				Path:       path,
			}

			fmt.Fprintf(out, formatter(p))
		})
	}
}

func formatter(p LogFormatterParams) string {
	statusColor := p.StatusCodeColor()
	methodColor := p.MethodColor()
	resetColor := p.ResetColor()

	if p.Latency > time.Minute {
		p.Latency = p.Latency - p.Latency%time.Second
	}

	return fmt.Sprintf("[MUX] %v |%s %3d %s| %13v |%s %-7s %s %s\n",
		p.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, p.StatusCode, resetColor,
		p.Latency,
		methodColor, p.Method, resetColor,
		p.Path,
	)
}
