package tcpproto

import (
	"fmt"
	"strings"
	"time"

	"github.com/Nigel2392/typeutils"
)

type Logger struct {
	Level string `json:"level"`
}

func (l *Logger) getMessage(t string, msg string, exclude []string) string {
	include := []string{"debug", "info", "warning", "error", "test"}
	if typeutils.Contains(exclude, t) {
		return ""
	}
	for _, v := range include {
		if v == t {
			return Colorize(l.GetLevelFromType(t), WrapTime(t, msg))
		}
	}

	return msg
}

func NewLogger(level string) *Logger {
	return &Logger{
		Level: level,
	}
}

func (l *Logger) Write(t string, msg string) {
	var console_msg string
	if l.GetLevel() >= 0 {
		console_msg = l.getMessage(t, msg, []string{})
	} else if l.GetLevel() >= 1 {
		console_msg = l.getMessage(t, msg, []string{"test"})
	} else if l.GetLevel() >= 2 {
		console_msg = l.getMessage(t, msg, []string{"test", "debug"})
	} else if l.GetLevel() >= 3 {
		console_msg = l.getMessage(t, msg, []string{"test", "debug", "info"})
	} else if l.GetLevel() >= 4 {
		console_msg = l.getMessage(t, msg, []string{"test", "debug", "info", "warning"})
	}
	fmt.Println(console_msg)
}

func (l *Logger) GetLevel() int {
	return l.GetLevelFromType(l.Level)
}
func (l *Logger) GetLevelFromType(t string) int {
	switch t {
	case "error":
		return 4
	case "warning":
		return 3
	case "info":
		return 2
	case "debug":
		return 1
	case "test":
		return 0
	default:
		return 1
	}
}

func (l *Logger) Info(msg string) {
	l.Write("info", msg)
}

func (l *Logger) Error(msg string) {
	l.Write("error", msg)
}

func (l *Logger) Warning(msg string) {
	l.Write("warning", msg)
}

func (l *Logger) Debug(msg string) {
	l.Write("debug", msg)
}

func (l *Logger) Test(msg string) {
	l.Write("test", msg)
}

func Colorize(level int, msg string) string {
	var (
		Reset  string = "\033[0m"
		Red    string = "\033[31m"
		Green  string = "\033[32m"
		Yellow string = "\033[33m"
		Blue   string = "\033[34m"
		Purple string = "\033[35m"
	)
	var selected string
	switch level {
	case 0:
		selected = Purple
	case 1:
		selected = Green
	case 2:
		selected = Blue
	case 3:
		selected = Yellow
	case 4:
		selected = Red
	default:
		selected = Green
	}
	return selected + msg + Reset
}
func WrapTime(t string, msg string) string {
	var time string = time.Now().Format("2006-01-02 15:04:05")
	var typ string = strings.ToUpper(t)
	return "[" + time + " " + typ + "] " + msg
}
