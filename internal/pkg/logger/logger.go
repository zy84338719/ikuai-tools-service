// Package logger is a thin bootstrap wrapper around github.com/oekit/logx.
// The oekit logx facade owns the actual logging (slog JSON backend with a
// rotating-file + stdout multi-sink); this package only wires the service
// config into it and exposes package-level helpers so call sites keep working.
package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/oekit/logx"
	"github.com/oekit/logx/sink"
)

// Config mirrors configs/config.yaml `log:` section.
type Config struct {
	Level      string `yaml:"level" json:"level"`           // debug|info|warn|error
	Filename   string `yaml:"filename" json:"filename"`     // empty = stdout only
	MaxSize    int    `yaml:"max_size" json:"max_size"`     // MB per file
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`       // days
	Compress   bool   `yaml:"compress" json:"compress"`
}

var multiSink *sink.MultiSink

// Init wires logx's default logger: slog JSON handler writing to
// stdout + a rotating file sink (when Filename is set). Safe to call once at
// bootstrap; later calls re-configure the default logger.
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{Level: "info"}
	}
	lvl := logx.ParseLogLevelOr(cfg.Level, logx.LevelInfo)

	slogLevel := slog.LevelInfo
	switch lvl {
	case logx.LevelDebug:
		slogLevel = slog.LevelDebug
	case logx.LevelWarn:
		slogLevel = slog.LevelWarn
	case logx.LevelError:
		slogLevel = slog.LevelError
	}

	var sinks []sink.Sink
	sinks = append(sinks, os.Stdout)
	if cfg.Filename != "" {
		rs, err := sink.NewRotatingSink(cfg.Filename, orDefault(cfg.MaxSize, 100),
			sink.WithMaxBackups(cfg.MaxBackups),
			sink.WithMaxAge(cfg.MaxAge),
			sink.WithCompress(cfg.Compress),
		)
		if err != nil {
			return fmt.Errorf("log file sink: %w", err)
		}
		sinks = append(sinks, rs)
	}
	multiSink = sink.NewMultiSink(sinks...)

	logx.SetDefault(logx.NewSlogLogger("ikuai-tools",
		logx.WithLevel(lvl),
		logx.WithHandler(slog.NewJSONHandler(multiSink, &slog.HandlerOptions{Level: slogLevel})),
	))
	return nil
}

// Sync flushes buffered sink writes. Call on shutdown.
func Sync() {
	if multiSink != nil {
		_ = multiSink.Sync()
	}
}

// Close flushes and closes the sinks.
func Close() {
	if multiSink != nil {
		_ = multiSink.Close()
	}
}

func Info(msg string, kv ...any) { logx.Info(msg, kv...) }
func Warn(msg string, kv ...any) { logx.Warn(msg, kv...) }
func Error(msg string, kv ...any) { logx.Error(msg, kv...) }
func Debug(msg string, kv ...any) { logx.Debug(msg, kv...) }

func orDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
