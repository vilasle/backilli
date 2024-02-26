package logger

import (
	"io"
	"log/slog"
	"os"
	"runtime"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var logger *slog.Logger
var currentLogger *slog.Logger

// enviroment - enviroment where application was started
// - local - text logger which is included debug level
// - dev - json logger which is included debug level
// - prod - json logger which is included info level
// if 'enviroment' has unexpected value will choose prod enviroment
// wrt - destination for writting logs. If it equals 'nil' then logger will write to os.Stdout

func Init(enviroment string, wrt io.Writer) {
	if wrt == nil {
		wrt = os.Stdout
	}

	switch enviroment {
	case envLocal:
		clearColorIfWindows()
		logger = slog.New(newHandler(wrt, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case enviroment:
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case enviroment:
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default: //default will use production enviroment
		logger = slog.New(slog.NewJSONHandler(wrt, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	currentLogger = logger
}

func With(args ...any) {
	currentLogger = logger.With(args...)
}

func Reset() {
	currentLogger = nil
}

func clearColorIfWindows() {
	if runtime.GOOS == "windows" {
		endcollor, magenta, blue, yellow, red, cyan, white =
			"", "", "", "", "", "", ""
	}
}

func Debug(format string, args ...any) {
	currentLogger.Debug(format, args...)
}

func Info(msg string, args ...any) {
	currentLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	currentLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	currentLogger.Error(msg, args...)
}
