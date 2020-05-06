package logging

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilteredLogging(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelInfo, out)

	assert.Equal(t, newLogger.level, LogLevel(2))
	assert.NotNil(t, newLogger.logger)

	newLogger.SetLogLevel(3)
	assert.Equal(t, newLogger.level, LogLevel(3))

	newLogger.Log(1, "this is hidden", "name")
	assert.Equal(t, "", out.String())
	out.Reset()

	newLogger.Log(4, "this is visible", "name")
	assert.Contains(t, out.String(), "this is visible")
	out.Reset()
}

func TestLogFormatting(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelInfo, out)

	newLogger.Log(LogLevelInfo, "test message", "test-name")
	assert.Contains(t, out.String(), "test message")
	assert.Contains(t, out.String(), "[Info]")
	assert.Contains(t, out.String(), "[test-name]")
	assert.Contains(t, out.String(), "[Flagship]")
}
