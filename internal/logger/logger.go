package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level 定义日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger 日志记录器
type Logger struct {
	level  Level
	logger *log.Logger
	format string // text 或 json
}

// New 创建新的日志记录器
func New(level Level, output io.Writer, format string) *Logger {
	if output == nil {
		output = os.Stdout
	}
	if format == "" {
		format = "text"
	}

	return &Logger{
		level:  level,
		logger: log.New(output, "", 0),
		format: format,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Debug 记录 debug 级别日志
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info 记录 info 级别日志
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn 记录 warn 级别日志
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error 记录 error 级别日志
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// log 内部日志记录方法
func (l *Logger) log(level Level, format string, args ...any) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var output string
	if l.format == "json" {
		output = fmt.Sprintf(`{"time":"%s","level":"%s","msg":"%s"}`,
			timestamp, level.String(), escapeJSON(msg))
	} else {
		output = fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), msg)
	}

	l.logger.Println(output)
}

// escapeJSON 转义 JSON 字符串
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// 全局默认日志记录器
var defaultLogger = New(LevelInfo, os.Stdout, "text")

// SetDefault 设置默认日志记录器
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// Debug 使用默认日志记录器记录 debug 日志
func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

// Info 使用默认日志记录器记录 info 日志
func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Warn 使用默认日志记录器记录 warn 日志
func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

// Error 使用默认日志记录器记录 error 日志
func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}
