package pgxc

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
	sctx "github.com/phathdt/service-context"
)

const (
	ansiReset        = "\033[0m"
	ansiBrightBlue   = "\033[94m"
	ansiBrightGreen  = "\033[92m"
	ansiBrightCyan   = "\033[96m"
	ansiBrightRed    = "\033[91m"
	ansiBrightYellow = "\033[93m"
)

type PgxLogAdapter struct {
	logger sctx.Logger
}

func colorizeQuery(msg string) string {
	msgLower := strings.ToLower(msg)
	if strings.Contains(msgLower, "select") {
		return ansiBrightBlue + msg + ansiReset
	} else if strings.Contains(msgLower, "insert") {
		return ansiBrightGreen + msg + ansiReset
	} else if strings.Contains(msgLower, "update") {
		return ansiBrightYellow + msg + ansiReset // Yellow
	} else if strings.Contains(msgLower, "delete") {
		return ansiBrightRed + msg + ansiReset // Red
	}
	return ansiBrightCyan + msg + ansiReset
}

func (l *PgxLogAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	// Skip if message contains "prepare" (case insensitive)
	if strings.Contains(strings.ToLower(msg), "prepare") {
		return
	}

	coloredMsg := colorizeQuery(msg)

	switch level {
	case tracelog.LogLevelTrace:
		l.logger.Debugf("%s %v", coloredMsg, data)
	case tracelog.LogLevelDebug:
		l.logger.Debugf("%s %v", coloredMsg, data)
	case tracelog.LogLevelInfo:
		l.logger.Infof("%s %v", coloredMsg, data)
	case tracelog.LogLevelWarn:
		l.logger.Warnf("%s %v", coloredMsg, data)
	case tracelog.LogLevelError:
		l.logger.Errorf("%s %v", coloredMsg, data)
	default:
		l.logger.Infof("%s %v", coloredMsg, data)
	}
}
