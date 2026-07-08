package main

import (
	"fmt"
	"os"

	"github.com/stevelan/cigarmender/internal/bamio"
	"github.com/stevelan/cigarmender/internal/cli"
	"github.com/stevelan/cigarmender/internal/log"
	"github.com/stevelan/cigarmender/internal/perf"
	"github.com/stevelan/cigarmender/internal/reference"

	"time"
)

func main() {

	defer perf.TimeTrack(time.Now(), "CIGARMender execution")

	log.SetupLogger(false)
	args := cli.ParseArgs()
	log.SetupLogger(args.Verbose)
	log.Info("\nStarted cigarmender")
	log.Verbose("Running cigarmender", "args", args)

	cli.CreateOutputDir(args.OutputDir)

	index := buildIndex(args)

	log.Info(index.Summary())

	bamVisitor := getProcessor(args, index)

	processBams(bamVisitor, args)

	log.Info(bamVisitor.Summary())

	log.Info(fmt.Sprintf("Mended BAMs written to : %s", args.OutputDir))

}

func buildIndex(args cli.Args) *reference.RefIndex {
	defer perf.TimeTrack(time.Now(), "Building index")
	hpindex, err := reference.IndexHomopolymers(args.Reference, args.HomopolymerSize, args.Bases)
	if err != nil {
		log.Error("Could not build index", "reference", args.Reference, "error", err)
	}
	return hpindex
}

func getProcessor(args cli.Args, index *reference.RefIndex) bamio.ReadProcessor {
	if args.DryRun {
		log.Info("Performing dry run, will read bam and report deletion statistics")
		return bamio.NewDelCounter()
	} else if args.Command == "readfilter" {
		return bamio.NewModCounter()
	} else {
		log.Verbose("Not dry run, centring deletions")
		log.Info("Rewriting bam file", "input", args.Input, "output", args.OutputDir)
		return bamio.NewDelCentrer(index)
	}
}

func processBams(bamVisitor bamio.ReadProcessor, args cli.Args) {
	defer perf.TimeTrack(time.Now(), "Processing BAMs")
	perfCapture := perf.Capture[int]{}

	count, err := perfCapture.CaptureCPU(func() (int, error) { return bamio.ReadBam(args.Input, bamVisitor, args) })

	if err != nil {
		log.Error("Error reading bam", "error", err)
		os.Exit(1)
	}
	log.Info("Processed reads", "readCount", count)
}
