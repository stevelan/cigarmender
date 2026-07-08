package log

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	LevelVerbose = slog.Level(-2)
)

// SetupLogger configure the logger with tint for pretty logs and set the log level based on the verbose flag
func SetupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		// change to Level.DEBUG in development
		level = LevelVerbose
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
	}))

	slog.SetDefault(logger)
}

// CloseAndLog - helper function that can be used with defer to close and resource and log any errors
func CloseAndLog(msg string, operation func() error) {
	err := operation()
	if err != nil {
		slog.Warn(msg, "error", err)
	}
}

// Verbose log a message at verbose levels, this is enabled with the verbose command line arg.
func Verbose(msg string, args ...any) {
	slog.Log(context.Background(), LevelVerbose, msg, args...)
}

// Delegate to slog for the other log levels
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Fatalf(message string, args ...any) {
	log.Fatalf("Could not build index for reference %s - %v", args...)
}
