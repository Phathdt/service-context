package pgxc

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
	sctx "github.com/phathdt/service-context"
)

const (
	ansiReset        = "\033[0m"
	ansiBrightBlue   = "\033[94m"
	ansiBrightGreen  = "\033[92m"
	ansiBrightYellow = "\033[93m"
	ansiBrightRed    = "\033[91m"
	ansiBrightCyan   = "\033[96m"
)

type PgxLogAdapter struct {
	logger sctx.Logger
}

// cleanSQL removes sqlc comments and minimizes SQL for logging
func cleanSQL(sql string) string {
	// Remove sqlc comments (-- name: ...)
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and sqlc comments
		if line == "" || strings.HasPrefix(line, "-- name:") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	
	// Join with spaces and clean up extra whitespace
	result := strings.Join(cleanLines, " ")
	
	// Remove multiple spaces
	re := regexp.MustCompile(`\s+`)
	result = re.ReplaceAllString(result, " ")
	
	return strings.TrimSpace(result)
}

// getSQLType determines the SQL operation type for coloring
func getSQLType(sql string) string {
	sqlLower := strings.ToLower(strings.TrimSpace(sql))
	
	if strings.HasPrefix(sqlLower, "select") {
		return "select"
	} else if strings.HasPrefix(sqlLower, "insert") {
		return "insert"
	} else if strings.HasPrefix(sqlLower, "update") {
		return "update"
	} else if strings.HasPrefix(sqlLower, "delete") {
		return "delete"
	} else if strings.HasPrefix(sqlLower, "create") || strings.HasPrefix(sqlLower, "alter") || strings.HasPrefix(sqlLower, "drop") {
		return "ddl"
	}
	
	return "other"
}

// colorizeSQL adds colors to SQL for text format only
func colorizeSQL(sql string, sqlType string, isTextFormat bool) string {
	if !isTextFormat {
		return sql
	}
	
	switch sqlType {
	case "select":
		return ansiBrightBlue + sql + ansiReset
	case "insert":
		return ansiBrightGreen + sql + ansiReset
	case "update":
		return ansiBrightYellow + sql + ansiReset
	case "delete":
		return ansiBrightRed + sql + ansiReset
	case "ddl":
		return ansiBrightCyan + sql + ansiReset
	default:
		return sql
	}
}

func (l *PgxLogAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	// Skip if message contains "prepare" (case insensitive)
	if strings.Contains(strings.ToLower(msg), "prepare") {
		return
	}

	// Detect if we're using text format for colorization
	isTextFormat := l.logger.GetFormat() != "json"

	// The actual SQL is in data["sql"], not in msg
	var actualSQL string
	var sqlType string = "other"
	
	if len(data) > 0 {
		if sqlStr, ok := data["sql"].(string); ok {
			actualSQL = cleanSQL(sqlStr)
			sqlType = getSQLType(actualSQL)
		}
	}

	// Colorize the message if it's "Query" and we have SQL
	displayMsg := msg
	if msg == "Query" && actualSQL != "" && isTextFormat {
		displayMsg = colorizeSQL(actualSQL, sqlType, isTextFormat)
	}

	// Use structured logging with Fields
	logger := l.logger
	if len(data) > 0 {
		// Clean SQL in data if present
		cleanedData := make(map[string]any)
		for k, v := range data {
			if k == "sql" {
				if sqlStr, ok := v.(string); ok {
					cleanedData[k] = cleanSQL(sqlStr)
				} else {
					cleanedData[k] = v
				}
			} else {
				cleanedData[k] = v
			}
		}
		
		// Add SQL type as metadata
		cleanedData["sql_type"] = sqlType
		logger = l.logger.Withs(sctx.Fields(cleanedData))
	} else {
		// Even without data, add sql_type
		logger = l.logger.Withs(sctx.Fields{"sql_type": sqlType})
	}

	// Just call the appropriate logger method
	switch level {
	case tracelog.LogLevelTrace:
		logger.Debug(displayMsg)
	case tracelog.LogLevelDebug:
		logger.Debug(displayMsg)
	case tracelog.LogLevelInfo:
		logger.Info(displayMsg)
	case tracelog.LogLevelWarn:
		logger.Warn(displayMsg)
	case tracelog.LogLevelError:
		logger.Error(displayMsg)
	default:
		logger.Info(displayMsg)
	}
}