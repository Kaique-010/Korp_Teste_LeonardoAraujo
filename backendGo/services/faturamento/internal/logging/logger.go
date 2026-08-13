package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const Version = "0.1.0"

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "debug",
	LevelInfo:  "info",
	LevelWarn:  "warn",
	LevelError: "error",
}

func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger emite logs estruturados em JSON (nível, timestamp, serviço, versão e campos).
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	level   Level
	service string
	version string
	fields  map[string]any
}

func New(service, version string) *Logger {
	return &Logger{
		out:     os.Stdout,
		level:   LevelInfo,
		service: service,
		version: version,
		fields:  map[string]any{},
	}
}

// With retorna um logger derivado com campos extras fixos.
func (l *Logger) With(fields map[string]any) *Logger {
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{out: l.out, level: l.level, service: l.service, version: l.version, fields: merged}
}

func (l *Logger) SetLevel(level Level) { l.level = level }

func (l *Logger) Debug(msg string, fields ...map[string]any) { l.log(LevelDebug, msg, fields...) }
func (l *Logger) Info(msg string, fields ...map[string]any)  { l.log(LevelInfo, msg, fields...) }
func (l *Logger) Warn(msg string, fields ...map[string]any)  { l.log(LevelWarn, msg, fields...) }
func (l *Logger) Error(msg string, fields ...map[string]any) { l.log(LevelError, msg, fields...) }

func (l *Logger) log(level Level, msg string, fields ...map[string]any) {
	if level < l.level {
		return
	}
	entry := map[string]any{
		"level":   levelNames[level],
		"time":    time.Now().UTC().Format(time.RFC3339),
		"service": l.service,
		"version": l.version,
		"message": msg,
	}
	for k, v := range l.fields {
		entry[k] = v
	}
	for _, f := range fields {
		for k, v := range f {
			entry[k] = v
		}
	}
	b, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.out, `{"level":"error","message":"falha ao serializar log: %v"}`+"\n", err)
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.out, string(b))
}

// std é o logger global usado por consumers e middleware; o serviço o define no main.
var std = New("korp", Version)

func SetDefault(l *Logger) { std = l }

func Debug(msg string, fields ...map[string]any) { std.Debug(msg, fields...) }
func Info(msg string, fields ...map[string]any)  { std.Info(msg, fields...) }
func Warn(msg string, fields ...map[string]any)  { std.Warn(msg, fields...) }
func Error(msg string, fields ...map[string]any) { std.Error(msg, fields...) }
