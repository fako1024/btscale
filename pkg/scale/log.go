package scale

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Logger denotes a generic log interface that logging service must provide
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

// NullLogger denotes a null-op logger that ignores all messages
type NullLogger struct{}

func (l *NullLogger) Error(args ...interface{}) {}

func (l *NullLogger) Errorf(format string, args ...interface{}) {}

func (l *NullLogger) Warn(args ...interface{}) {}

func (l *NullLogger) Warnf(format string, args ...interface{}) {}

func (l *NullLogger) Info(args ...interface{}) {}

func (l *NullLogger) Infof(format string, args ...interface{}) {}

func (l *NullLogger) Debug(args ...interface{}) {}

func (l *NullLogger) Debugf(format string, args ...interface{}) {}

// NewDefaultLogger instantiates a new default logger
func NewDefaultLogger(debug bool) *zap.SugaredLogger {

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	logCfg.DisableCaller = level > zap.DebugLevel
	logCfg.Level.SetLevel(level)
	zapLogger, err := logCfg.Build()
	if err != nil {
		fmt.Printf("failed to instantiate logger: %s\n", err)
		os.Exit(1)
	}

	return zapLogger.Sugar()
}
