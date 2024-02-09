package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

const (
	envLocal      = "local"
	envProduction = "prod"
	envDebug      = "debug"
)

var logger *slog.Logger

// env - type of enviroment
// wrt - where need to write logs
func InitLogger(env string, wrt io.Writer) {
	if wrt == nil {
		wrt = os.Stdout
	}
	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(wrt, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDebug:
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProduction:
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default: //default will use production enviroment
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Errorf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}
