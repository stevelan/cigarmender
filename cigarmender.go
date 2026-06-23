package main

import (
	"cigarmender/args"
	"cigarmender/bamreader"
	"cigarmender/reference"
	"log"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	slog.Info("Task completed", "task", name, "elapsed", elapsed)
}

func main() {
	defer timeTrack(time.Now(), "CIGARMender execution")

	args := args.ParseArgs()
	setupLogger(args.Verbose)

	slog.Info("Running cigarmender", "args", args)
	log.Printf("Processing cigarmender with %s", args.String())
	index := buildIndex(args)
	slog.Info(index.Summary())

	bamVisitor := getVisitor(args)
	count, err := bamreader.ReadBam(args.Input, bamVisitor, index, args)
	if err != nil {
		log.Fatalf("Error reading bam %v", err)
	}
	log.Printf("Processed %d reads", count)

	log.Println(bamVisitor.Summary())
}

func buildIndex(args args.Args) *reference.RefIndex {
	defer timeTrack(time.Now(), "Building index")
	hpindex, err := reference.IndexHomopolymers(args.Reference, args.HomopolymerSize, args.Bases)
	if err != nil {
		log.Fatalf("Could not build index for reference %s - %v", args.Reference, err)
	}
	return hpindex
}

func getVisitor(args args.Args) bamreader.ReadVisitor {
	if args.DryRun {
		log.Println("Performing dry run, will read bam and report deletion statistics")
		return bamreader.NewDelCounter()
	}

	log.Fatalf("Implement me")
	return nil
}

func setupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
	}))

	slog.SetDefault(logger)
}
