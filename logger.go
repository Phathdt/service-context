package sctx

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type Fields map[string]any

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Fatal(...any)
	Panic(...any)
	Trace(...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)
	Tracef(format string, args ...any)

	Debugln(...any)
	Infoln(...any)
	Warnln(...any)
	Errorln(...any)
	Fatalln(...any)
	Panicln(...any)
	Traceln(...any)

	With(key string, value any) Logger
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

func (l *logger) log(level CustomLevel, args ...any) {
	if !l.Logger.Enabled(context.Background(), level.Level()) {
		return
	}
	msg := fmt.Sprint(args...)
	l.Logger.Log(context.Background(), level.Level(), msg)
}

func (l *logger) GetSLogger() *slog.Logger {
	return l.Logger
}
func (l *logger) Debug(args ...any) { l.debugSrc().log(LevelDebug, args...) }
func (l *logger) Info(args ...any)  { l.log(LevelInfo, args...) }
func (l *logger) Warn(args ...any)  { l.log(LevelWarn, args...) }
func (l *logger) Error(args ...any) { l.log(LevelError, args...) }
func (l *logger) Fatal(args ...any) { l.log(LevelFatal, args...); os.Exit(1) }
func (l *logger) Panic(args ...any) { s := fmt.Sprint(args...); l.log(LevelPanic, s); panic(s) }
func (l *logger) Trace(args ...any) { l.log(LevelTrace, args...) }

func (l *logger) logf(level CustomLevel, format string, args ...any) {
	if !l.Logger.Enabled(context.Background(), level.Level()) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.Logger.Log(context.Background(), level.Level(), msg)
}

func (l *logger) Debugf(format string, args ...any) {
	l.debugSrc().logf(LevelDebug, format, args...)
}
func (l *logger) Infof(format string, args ...any)  { l.logf(LevelInfo, format, args...) }
func (l *logger) Warnf(format string, args ...any)  { l.logf(LevelWarn, format, args...) }
func (l *logger) Errorf(format string, args ...any) { l.logf(LevelError, format, args...) }
func (l *logger) Fatalf(format string, args ...any) {
	l.logf(LevelFatal, format, args...)
	os.Exit(1)
}
func (l *logger) Panicf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	l.logf(LevelPanic, "%s", s)
	panic(s)
}
func (l *logger) Tracef(format string, args ...any) { l.logf(LevelTrace, format, args...) }

func (l *logger) Debugln(args ...any) { l.Debug(args...) }
func (l *logger) Infoln(args ...any)  { l.Info(args...) }
func (l *logger) Warnln(args ...any)  { l.Warn(args...) }
func (l *logger) Errorln(args ...any) { l.Error(args...) }
func (l *logger) Fatalln(args ...any) {
	l.Fatal(args...)
	// Note: os.Exit(1) is called by l.Fatal, so it's not needed here
}
func (l *logger) Panicln(args ...any) {
	l.Panic(args...)
	// Note: panic is called by l.Panic, so it's not needed here
}
func (l *logger) Traceln(args ...any) { l.Trace(args...) }

func (l *logger) With(key string, value any) Logger {
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

	l := createSlogLogger(mustParseLevel(config.DefaultLevel))

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
	al.logger = createSlogLogger(mustParseLevel(al.logLevel))

	return nil
}

func (al *appLogger) Stop() error {
	return nil
}

const (
	ansiReset          = "\033[0m"
	ansiFaint          = "\033[2m"
	ansiResetFaint     = "\033[22m"
	ansiBrightRed      = "\033[91m"
	ansiBrightGreen    = "\033[92m"
	ansiBrightYellow   = "\033[93m"
	ansiBrightBlue     = "\033[94m"
	ansiBrightMagenta  = "\033[95m"
	ansiBrightCyan     = "\033[96m"
	ansiBrightRedFaint = "\033[91;2m"
	ansiBackgroundRed  = "\033[41m"
)

func createSlogLogger(level CustomLevel) *slog.Logger {
	w := os.Stderr
	return slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      level.Level(),
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.DateTime,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					lvl := a.Value.Any().(slog.Level)
					switch {
					case lvl == LevelTrace.Level():
						a.Value = slog.StringValue(ansiFaint + "TRACE" + ansiReset)
					case lvl == LevelDebug.Level():
						a.Value = slog.StringValue(ansiBrightBlue + "DEBUG" + ansiReset)
					case lvl == LevelInfo.Level():
						a.Value = slog.StringValue(ansiBrightGreen + "INFO" + ansiReset)
					case lvl == LevelWarn.Level():
						a.Value = slog.StringValue(ansiBrightYellow + "WARN" + ansiReset)
					case lvl == LevelError.Level():
						a.Value = slog.StringValue(ansiBrightRed + "ERROR" + ansiReset)
					case lvl == LevelFatal.Level():
						a.Value = slog.StringValue(ansiBrightMagenta + "FATAL" + ansiReset)
					case lvl == LevelPanic.Level():
						a.Value = slog.StringValue(ansiBackgroundRed + ansiBrightCyan + "PANIC" + ansiReset)
					default:
						a.Value = slog.StringValue("UNKNOWN")
					}
				}
				return a
			},
		}),
	)
}
