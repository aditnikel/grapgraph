package observability

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger struct {
	level string
}

func New(level string) *Logger {
	return &Logger{level: normalizeLevel(level)}
}

type Fields map[string]any

func (l *Logger) Debug(msg string, f Fields) { l.log("debug", msg, f) }
func (l *Logger) Info(msg string, f Fields)  { l.log("info", msg, f) }
func (l *Logger) Warn(msg string, f Fields)  { l.log("warn", msg, f) }
func (l *Logger) Error(msg string, f Fields) { l.log("error", msg, f) }

func (l *Logger) log(level, msg string, f Fields) {
	if l == nil {
		return
	}

	level = normalizeLevel(level)
	current := levelPriority(l.level)
	incoming := levelPriority(level)
	if incoming < current {
		return
	}

	m := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}
	if f != nil {
		for k, v := range f {
			m[k] = v
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"ts":"%s","level":"error","msg":"log_marshal_failed","err":"%v"}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), err)
		return
	}
	fmt.Fprintln(os.Stdout, string(b))
}

func normalizeLevel(level string) string {
	switch level {
	case "debug", "info", "warn", "error":
		return level
	default:
		return "info"
	}
}

func levelPriority(level string) int {
	switch level {
	case "debug":
		return 10
	case "info":
		return 20
	case "warn":
		return 30
	case "error":
		return 40
	default:
		return 20
	}
}
