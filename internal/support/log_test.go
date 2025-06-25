package support_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/trento-project/workbench/internal/support"
)

func NewDefaultLoggerMock(w io.Writer) *slog.Logger {
	return slog.New(support.NewDefaultTextHandler(w, slog.LevelInfo))
}

func TestDefaultTextHandlerInfoLog(t *testing.T) {
	var buf bytes.Buffer

	logger := NewDefaultLoggerMock(&buf)

	logger.Info("This is an info message")

	expected := "INFO This is an info message\n"

	actual := stripTimestamp(buf.String())

	if actual != expected {
		t.Errorf("expected log line %q, got %q", expected, actual)
	}

}

func TestDefaultTextHandlerInfoLogWithAttr(t *testing.T) {
	var buf bytes.Buffer

	logger := NewDefaultLoggerMock(&buf)

	logger.Info("This is an info message", "my_attr", "my_value")

	expected := "INFO This is an info message my_attr=my_value\n"

	actual := stripTimestamp(buf.String())

	if actual != expected {
		t.Errorf("expected log line %q, got %q", expected, actual)
	}

}

func TestDefaultTextHandlerInfoLogWithDefaultAttr(t *testing.T) {
	var buf bytes.Buffer

	logger := NewDefaultLoggerMock(&buf).With("default_attr", "default_value")

	logger.Info("This is an info message", "my_attr", "my_value")

	expected := "INFO This is an info message my_attr=my_value default_attr=default_value\n"

	actual := stripTimestamp(buf.String())

	if actual != expected {
		t.Errorf("expected log line %q, got %q", expected, actual)
	}

}

func TestDefaultTextHandlerInfoLogWithGroupAndDefaultAttr(t *testing.T) {
	var buf bytes.Buffer

	logger := NewDefaultLoggerMock(&buf).
		With("attr_a", "value_a").
		WithGroup("group_x").
		With("attr_b", "value_b").
		WithGroup("group_y")

	logger.Info("This is an info message", "my_attr", "my_value")

	expected := "INFO This is an info message group_x.group_y.my_attr=my_value attr_a=value_a group_x.attr_b=value_b\n"

	actual := stripTimestamp(buf.String())

	if actual != expected {
		t.Errorf("expected log line %q, got %q", expected, actual)
	}

}

// Helper to strip the timestamp from the log line
func stripTimestamp(line string) string {
	// Timestamp is always 19 chars: "YYYY-MM-DD hh:mm:ss"
	if len(line) > 20 {
		return line[20:]
	}
	return ""
}
