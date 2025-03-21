package plugininterface

import (
	"testing"
)

func TestPluginError(t *testing.T) {
	t.Run("ErrorString", func(t *testing.T) {
		err := &PluginError{
			Name:    "test",
			Message: "test error",
		}
		expected := "plugin test: test error"
		if err.Error() != expected {
			t.Errorf("Expected error string %q, got %q", expected, err.Error())
		}
	})

	t.Run("ErrorStringWithCause", func(t *testing.T) {
		cause := &PluginError{
			Name:    "cause",
			Message: "cause error",
		}
		err := &PluginError{
			Name:    "test",
			Message: "test error",
			Cause:   cause,
		}
		expected := "plugin test: test error: plugin cause: cause error"
		if err.Error() != expected {
			t.Errorf("Expected error string %q, got %q", expected, err.Error())
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := &PluginError{
			Name:    "cause",
			Message: "cause error",
		}
		err := &PluginError{
			Name:    "test",
			Message: "test error",
			Cause:   cause,
		}
		if err.Unwrap() != cause {
			t.Error("Expected Unwrap to return the cause error")
		}
	})
}
