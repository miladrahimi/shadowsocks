package logger

import (
	"errors"
	"fmt"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"syscall"
)

type Logger struct {
	Engine *zap.Logger
}

func (l *Logger) Shutdown() {
	if err := l.Engine.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		l.Engine.Warn("cannot close the log", zap.Error(err))
	} else {
		l.Engine.Debug("log closed successfully")
	}
}

func New(c *config.Config) (logger *Logger, err error) {
	level := zap.NewAtomicLevel()
	if err = level.UnmarshalText([]byte(c.Logger.Level)); err != nil {
		return nil, errors.New(fmt.Sprintf("invalid log level %s, err: %v", c.Logger.Level, err))
	}

	engine, err := zap.Config{
		Level:             level,
		Development:       false,
		Encoding:          "json",
		DisableStacktrace: true,
		DisableCaller:     true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			EncodeTime:     zapcore.TimeEncoderOfLayout(c.Logger.Format),
			EncodeDuration: zapcore.StringDurationEncoder,
			LevelKey:       "level",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			NameKey:        "key",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			LineEnding:     zapcore.DefaultLineEnding,
		},
	}.Build()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot build logger, err: %v", err))
	}

	return &Logger{engine}, nil
}
