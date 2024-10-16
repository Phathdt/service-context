package sctx

import (
	"context"
	"flag"
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type Fields map[string]any

type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Panic(...interface{})
	Trace(...interface{})

	With(key string, value interface{}) Logger
	Withs(Fields) Logger
	WithSrc() Logger
	GetLevel() string

	GetSLogger() *slog.Logger
}

type CustomLevel int

const (
	LevelTrace CustomLevel = iota - 1
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

func (l CustomLevel) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	default:
		return fmt.Sprintf("level(%d)", l)
	}
}

func (l CustomLevel) Level() slog.Level {
	switch l {
	case LevelTrace:
		return slog.LevelDebug - 1
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 1
	case LevelPanic:
		return slog.LevelError + 2
	default:
		return slog.LevelInfo
	}
}

type logger struct {
	*slog.Logger
	level CustomLevel
}

func (l *logger) GetLevel() string {
	return l.level.String()
}

func (l *logger) debugSrc() *logger {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}
	return &logger{l.Logger.With("source", fmt.Sprintf("%s:%d", file, line)), l.level}
}

func (l *logger) log(level CustomLevel, args ...interface{}) {
	if !l.Logger.Enabled(context.Background(), level.Level()) {
		return
	}
	msg := fmt.Sprint(args...)
	l.Logger.Log(context.Background(), level.Level(), msg)
}

func (l *logger) GetSLogger() *slog.Logger {
	return l.Logger
}
func (l *logger) Debug(args ...interface{}) { l.debugSrc().log(LevelDebug, args...) }
func (l *logger) Info(args ...interface{})  { l.log(LevelInfo, args...) }
func (l *logger) Warn(args ...interface{})  { l.log(LevelWarn, args...) }
func (l *logger) Error(args ...interface{}) { l.log(LevelError, args...) }
func (l *logger) Fatal(args ...interface{}) { l.log(LevelFatal, args...); os.Exit(1) }
func (l *logger) Panic(args ...interface{}) { s := fmt.Sprint(args...); l.log(LevelPanic, s); panic(s) }
func (l *logger) Trace(args ...interface{}) { l.log(LevelTrace, args...) }

func (l *logger) With(key string, value interface{}) Logger {
	return &logger{l.Logger.With(key, value), l.level}
}

func (l *logger) Withs(fields Fields) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return &logger{l.Logger.With(attrs...), l.level}
}

func (l *logger) WithSrc() Logger {
	return l.debugSrc()
}

func mustParseLevel(level string) CustomLevel {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "panic":
		return LevelPanic
	default:
		panic(fmt.Sprintf("invalid log level: %s", level))
	}
}

var (
	defaultLogger = newAppLogger(&Config{
		BasePrefix:   "core",
		DefaultLevel: "debug",
	})
)

type AppLogger interface {
	GetLogger(prefix string) Logger
}

func GlobalLogger() AppLogger {
	return defaultLogger
}

type Config struct {
	DefaultLevel string
	BasePrefix   string
}

type appLogger struct {
	logger   *slog.Logger
	cfg      Config
	logLevel string
}

func newAppLogger(config *Config) *appLogger {
	if config == nil {
		config = &Config{}
	}

	if config.DefaultLevel == "" {
		config.DefaultLevel = "info"
	}

	w := os.Stderr
	l := slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      mustParseLevel(config.DefaultLevel).Level(),
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.DateTime,
		}),
	)
	return &appLogger{
		logger:   l,
		cfg:      *config,
		logLevel: config.DefaultLevel,
	}
}

func (al *appLogger) GetLogger(prefix string) Logger {
	prefix = al.cfg.BasePrefix + "." + prefix
	prefix = strings.Trim(prefix, ".")

	l := al.logger
	if prefix != "" {
		l = l.With("prefix", prefix)
	}

	return &logger{l, mustParseLevel(al.logLevel)}
}

func (*appLogger) ID() string {
	return "logger"
}

func (al *appLogger) InitFlags() {
	flag.StringVar(&al.logLevel, "log-level", al.cfg.DefaultLevel, "Log level: trace | debug | info | warn | error | fatal | panic")
}

func (al *appLogger) Activate(_ ServiceContext) error {
	w := os.Stderr

	al.logger = slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      mustParseLevel(al.logLevel).Level(),
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.DateTime,
		}),
	)

	return nil
}

func (al *appLogger) Stop() error {
	return nil
}
