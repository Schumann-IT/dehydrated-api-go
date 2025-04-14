// Package logger provides a centralized logging facility for the dehydrated-api-go application.
// It wraps the zap logger and provides a simple interface for configuring and using logging
// throughout the application.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// globalLogger is the singleton instance of the zap logger used throughout the application.
	// It is initialized with default settings if not explicitly configured.
	globalLogger *zap.Logger
)

// Config holds the configuration for the logger.
// It allows customization of log level, output format, and destination.
type Config struct {
	// Level specifies the minimum log level to output (debug, info, warn, error).
	Level string `yaml:"level"`

	// Encoding specifies the output format (json or console).
	Encoding string `yaml:"encoding"`

	// OutputPath specifies the path to the log file. If empty, logs are written to stdout.
	OutputPath string `yaml:"outputPath"`
}

// DefaultConfig returns a new Config with default settings.
// The log level is determined by the LOG_LEVEL environment variable,
// defaulting to "error" if not set. The encoding defaults to "console".
func DefaultConfig() *Config {
	l := os.Getenv("LOG_LEVEL")
	if l == "" {
		l = "error"
	}
	return &Config{
		Level:    l,
		Encoding: "console",
	}
}

// Init initializes the global logger with the given configuration.
// It sets up the logger with the specified level, encoding, and output destination.
// Returns an error if the configuration is invalid or if there are issues creating the log file.
func Init(cfg *Config) error {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder based on encoding type
	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create output writer
	var output zapcore.WriteSyncer
	if cfg.OutputPath != "" {
		output, err = os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		output,
		level,
	)

	// Create logger
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// L returns the global logger instance.
// If the logger hasn't been initialized, it will be initialized with default settings.
// This function is the primary way to access the logger throughout the application.
func L() *zap.Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		Init(DefaultConfig())
	}
	return globalLogger
}

// Sync flushes any buffered log entries and closes any open log files.
// This function should be called before the application exits to ensure
// all log entries are properly written.
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
