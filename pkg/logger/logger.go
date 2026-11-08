// Package logger 提供结构化日志支持，按级别过滤、支持文本与 JSON 格式
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// =====================================================================
// 级别定义
// =====================================================================

// Level 日志级别
type Level int32

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String 返回级别字符串
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
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 从字符串解析级别
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	case "fatal", "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Format 输出格式
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// ParseFormat 从字符串解析格式
func ParseFormat(s string) Format {
	if s == "json" || s == "JSON" {
		return FormatJSON
	}
	return FormatText
}

// =====================================================================
// Logger 实现
// =====================================================================

// Field 结构化日志字段
type Field struct {
	Key   string
	Value interface{}
}

// F 创建一个 Field
func F(key string, value interface{}) Field { return Field{Key: key, Value: value} }

// Logger 结构化日志记录器
type Logger struct {
	level  atomic.Int32
	format atomic.Int32
	out    io.Writer
	mu     sync.Mutex
	fields []Field
}

// New 创建一个 Logger
func New(level Level, format Format, out io.Writer) *Logger {
	l := &Logger{out: out}
	l.level.Store(int32(level))
	l.format.Store(int32(format))
	return l
}

// Default 全局默认 logger
var Default = New(LevelInfo, FormatText, os.Stdout)

// SetLevel 设置级别
func (l *Logger) SetLevel(level Level) { l.level.Store(int32(level)) }

// SetFormat 设置格式
func (l *Logger) SetFormat(format Format) { l.format.Store(int32(format)) }

// SetOutput 设置输出目标
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

// With 返回带有公共字段的子 logger
func (l *Logger) With(fields ...Field) *Logger {
	combined := make([]Field, 0, len(l.fields)+len(fields))
	combined = append(combined, l.fields...)
	combined = append(combined, fields...)
	return &Logger{
		out:    l.out,
		fields: combined,
		// 复制原子状态
	}
}

// Level 获取当前级别
func (l *Logger) Level() Level { return Level(l.level.Load()) }

// =====================================================================
// 日志方法
// =====================================================================

// Debug 记录 DEBUG 级别日志
func (l *Logger) Debug(msg string, fields ...Field) { l.log(LevelDebug, msg, fields) }

// Info 记录 INFO 级别日志
func (l *Logger) Info(msg string, fields ...Field) { l.log(LevelInfo, msg, fields) }

// Warn 记录 WARN 级别日志
func (l *Logger) Warn(msg string, fields ...Field) { l.log(LevelWarn, msg, fields) }

// Error 记录 ERROR 级别日志
func (l *Logger) Error(msg string, fields ...Field) { l.log(LevelError, msg, fields) }

// Errorf 记录 ERROR 级别日志，支持格式化
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...), nil)
}

// Infof 记录 INFO 级别日志，支持格式化
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...), nil)
}

// Warnf 记录 WARN 级别日志，支持格式化
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...), nil)
}

// Debugf 记录 DEBUG 级别日志，支持格式化
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...), nil)
}

// Fatal 记录 FATAL 级别日志后调用 os.Exit(1)
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(LevelFatal, msg, fields)
	os.Exit(1)
}

// Fatalf 记录 FATAL 级别日志后调用 os.Exit(1)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(LevelFatal, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

// log 写入一条日志
func (l *Logger) log(level Level, msg string, fields []Field) {
	if level < l.Level() {
		return
	}
	ts := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	combined := make([]Field, 0, len(l.fields)+len(fields))
	combined = append(combined, l.fields...)
	combined = append(combined, fields...)

	var line string
	if Format(l.format.Load()) == FormatJSON {
		line = formatJSON(level, ts, msg, combined)
	} else {
		line = formatText(level, ts, msg, combined)
	}

	l.mu.Lock()
	_, _ = io.WriteString(l.out, line+"\n")
	l.mu.Unlock()
}

// formatText 文本格式
// 2006-01-02T15:04:05.000Z07:00 INFO message key=value key2=value2
func formatText(level Level, ts, msg string, fields []Field) string {
	var b []byte
	b = append(b, ts...)
	b = append(b, ' ')
	b = append(b, level.String()...)
	b = append(b, ' ')
	b = append(b, msg...)
	for _, f := range fields {
		b = append(b, ' ')
		b = append(b, f.Key...)
		b = append(b, '=')
		b = append(b, fmt.Sprintf("%v", f.Value)...)
	}
	return string(b)
}

// formatJSON JSON 格式（单行）
func formatJSON(level Level, ts, msg string, fields []Field) string {
	out := map[string]interface{}{
		"timestamp": ts,
		"level":     level.String(),
		"message":   msg,
	}
	for _, f := range fields {
		out[f.Key] = f.Value
	}
	// 简单 JSON 编码（使用 fmt 处理常见类型）
	var sb []byte
	sb = append(sb, '{')
	sb = append(sb, fmt.Sprintf(`"timestamp":"%s","level":"%s","message":%q`, ts, level.String(), msg)...)
	for _, f := range fields {
		sb = append(sb, ',')
		sb = append(sb, fmt.Sprintf(`%q:`, f.Key)...)
		sb = append(sb, encodeValue(f.Value)...)
	}
	sb = append(sb, '}')
	return string(sb)
}

// encodeValue 简单 JSON 值编码
func encodeValue(v interface{}) []byte {
	switch t := v.(type) {
	case string:
		return []byte(fmt.Sprintf("%q", t))
	case error:
		return []byte(fmt.Sprintf("%q", t.Error()))
	case fmt.Stringer:
		return []byte(fmt.Sprintf("%q", t.String()))
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
