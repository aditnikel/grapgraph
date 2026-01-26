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
	return &Logger{level: level}
}

type Fields map[string]any

func (l *Logger) Info(msg string, f Fields)  { l.log("info", msg, f) }
func (l *Logger) Warn(msg string, f Fields)  { l.log("warn", msg, f) }
func (l *Logger) Error(msg string, f Fields) { l.log("error", msg, f) }

func (l *Logger) log(level, msg string, f Fields) {
	if l.level == "error" && level != "error" {
		return
	}
	if l.level == "warn" && level == "info" {
		return
	}

	m := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}
	for k, v := range f {
		m[k] = v
	}

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"ts":"%s","level":"error","msg":"log_marshal_failed","err":"%v"}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), err)
		return
	}
	fmt.Fprintln(os.Stdout, string(b))
}
