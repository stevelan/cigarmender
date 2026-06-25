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

func Verbose(msg string, args ...any) {
	slog.Log(context.Background(), LevelVerbose, msg, args...)
}

func Fatalf(message string, args ...any) {
	log.Fatalf("Could not build index for reference %s - %v", args...)
}

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
