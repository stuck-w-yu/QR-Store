package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var Log *slog.Logger

type PrettyHandler struct {
	opts slog.HandlerOptions
	out  io.Writer
	mu   *sync.Mutex
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts: *opts,
		out:  out,
		mu:   &sync.Mutex{},
	}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("15:04:05")

	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = "\033[90mDEBUG\033[0m"
	case slog.LevelInfo:
		levelStr = "\033[36mINFO \033[0m"
	case slog.LevelWarn:
		levelStr = "\033[33mWARN \033[0m"
	case slog.LevelError:
		levelStr = "\033[31mERROR\033[0m"
	default:
		levelStr = r.Level.String()
	}

	attrs := make([]string, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("\033[90m%s=\033[0m%v", a.Key, a.Value.Any()))
		return true
	})

	var attrStr string
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(h.out, "\033[90m%s\033[0m  %s  %s%s\n", timeStr, levelStr, r.Message, attrStr)
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init(env string) {
	var handler slog.Handler

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = NewPrettyHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// LogHTTP prints a clean, formatted HTTP access log line
func LogHTTP(status int, method, path string, latency time.Duration, clientIP string) {
	timeStr := time.Now().Format("15:04:05")

	var statusColor string
	switch {
	case status >= 200 && status < 300:
		statusColor = fmt.Sprintf("\033[32m%d\033[0m", status)
	case status >= 300 && status < 400:
		statusColor = fmt.Sprintf("\033[36m%d\033[0m", status)
	case status >= 400 && status < 500:
		statusColor = fmt.Sprintf("\033[33m%d\033[0m", status)
	default:
		statusColor = fmt.Sprintf("\033[31m%d\033[0m", status)
	}

	var methodColor string
	switch method {
	case "GET":
		methodColor = "\033[34mGET   \033[0m"
	case "POST":
		methodColor = "\033[32mPOST  \033[0m"
	case "PATCH":
		methodColor = "\033[33mPATCH \033[0m"
	case "PUT":
		methodColor = "\033[33mPUT   \033[0m"
	case "DELETE":
		methodColor = "\033[31mDELETE\033[0m"
	case "OPTIONS":
		methodColor = "\033[90mOPTION\033[0m"
	default:
		methodColor = fmt.Sprintf("%-6s", method)
	}

	var latencyStr string
	if latency < time.Millisecond {
		latencyStr = fmt.Sprintf("%6.0fµs", float64(latency.Microseconds()))
	} else if latency < time.Second {
		latencyStr = fmt.Sprintf("%6.2fms", float64(latency.Microseconds())/1000.0)
	} else {
		latencyStr = fmt.Sprintf("%6.2fs", latency.Seconds())
	}

	fmt.Printf("\033[90m%s\033[0m  \033[35mHTTP \033[0m  %s │ %s │ %-15s │ %s %s\n",
		timeStr, statusColor, latencyStr, clientIP, methodColor, path)
}

// PrintBanner prints a clean server startup banner
func PrintBanner(port, env, dbStatus, corsOrigins string) {
	width := 72
	line := strings.Repeat("─", width)

	fmt.Println()
	fmt.Printf("\033[1;36m┌%s┐\033[0m\n", line)

	title := "QR-STORE BACKEND SERVER"
	padLeft := (width - len(title)) / 2
	padRight := width - len(title) - padLeft
	fmt.Printf("\033[1;36m│\033[0m%s\033[1;37m%s\033[0m%s\033[1;36m│\033[0m\n", strings.Repeat(" ", padLeft), title, strings.Repeat(" ", padRight))
	fmt.Printf("\033[1;36m├%s┤\033[0m\n", line)

	printRow := func(label, value string) {
		plainText := fmt.Sprintf("  %-10s : %s", label, value)
		if len(plainText) > width {
			plainText = plainText[:width-3] + "..."
		}
		padding := width - len(plainText)
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("\033[1;36m│\033[0m\033[1m  %-10s\033[0m : %s%s\033[1;36m│\033[0m\n", label, value, strings.Repeat(" ", padding))
	}

	printRow("Status", "● Ready")
	printRow("Port", fmt.Sprintf("%s (http://localhost:%s)", port, port))
	printRow("Env", env)
	printRow("Database", dbStatus)
	printRow("CORS", corsOrigins)

	fmt.Printf("\033[1;36m└%s┘\033[0m\n", line)
	fmt.Println()
}
