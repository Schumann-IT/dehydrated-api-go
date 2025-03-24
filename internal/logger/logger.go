package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	globalLogger *zap.Logger
)

// Config holds logging configuration
type Config struct {
	Level      string `yaml:"level"`      // debug, info, warn, error
	Encoding   string `yaml:"encoding"`   // json or console
	OutputPath string `yaml:"outputPath"` // path to log file, empty for stdout
}

// DefaultConfig returns default logging configuration
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

// Init initializes the global logger with the given configuration
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

// L returns the global logger instance
func L() *zap.Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		Init(DefaultConfig())
	}
	return globalLogger
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
