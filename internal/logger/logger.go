package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// defaultLoggerConfig returns a new Config with default settings.
// The log level is determined by the LOG_LEVEL environment variable,
// defaulting to "error" if not set. The encoding defaults to "console".
func DefaultLoggerConfig() *Config {
	l := os.Getenv("LOG_LEVEL")
	if l == "" {
		l = "error"
	}
	return &Config{
		Level:    l,
		Encoding: "console",
	}
}

func NewLogger(cfg *Config) (*zap.Logger, error) {
	if cfg == nil {
		cfg = DefaultLoggerConfig()
	}
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
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
			return nil, err
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

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
