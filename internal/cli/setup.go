package cli

import (
	"log/slog"
	"os"

	"github.com/stevelan/cigarmender/internal/log"
)

func CreateOutputDir(outdir string) {
	log.Verbose("Creating output directory if it does not exist", "output", outdir)
	err := os.MkdirAll(outdir, os.ModePerm)
	if err != nil {
		slog.Error("Could not create output directory", "output", outdir, "error", err)
		os.Exit(1)
	}
}
