package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger for structured logging
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger instance
func New() *Logger {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	var level zapcore.Level

	env := os.Getenv("ENVIRONMENT")
	logLevel := os.Getenv("LOG_LEVEL")

	// Parse log level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Development mode uses console encoder
	if env == "development" {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		encoder = zapcore.NewJSONEncoder(config)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{logger.Sugar()}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields ...interface{}) *Logger {
	return &Logger{l.With(fields...)}
}

// WithError creates a new logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.With("error", err)}
}

// WithRequestID creates a new logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{l.With("request_id", requestID)}
}

// WithTenantID creates a new logger with a tenant ID field
func (l *Logger) WithTenantID(tenantID string) *Logger {
	return &Logger{l.With("tenant_id", tenantID)}
}

// WithUserID creates a new logger with a user ID field
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{l.With("user_id", userID)}
}

// WithAgentID creates a new logger with an agent ID field
func (l *Logger) WithAgentID(agentID string) *Logger {
	return &Logger{l.With("agent_id", agentID)}
}

