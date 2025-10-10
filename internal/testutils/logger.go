package testutils

import (
	"github.com/cristianoliveira/aerospace-scratchpad/internal/logger"
)

type TestingLogger struct {
	Logger func(msg string, args ...any)
}

func (l *TestingLogger) LogInfo(msg string, args ...any) {
	l.Logger(msg, args...)
}
func (l *TestingLogger) LogError(msg string, args ...any) {
	l.Logger(msg, args...)
}
func (l *TestingLogger) LogDebug(msg string, args ...any) {
	l.Logger(msg, args...)
}
func (l *TestingLogger) Close() error {
	// No-op
	return nil
}
func (l *TestingLogger) GetConfig() logger.LogConfig {
	// No-op
	return logger.LogConfig{
		Path:  "/tmp/aerospace-marks.log",
		Level: "DISABLED",
	}
}
func (l *TestingLogger) AsJSON(_ any) string {
	// No-op
	return ""
}
