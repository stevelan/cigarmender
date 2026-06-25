package main

import (
	"cigarmender/args"
	"cigarmender/bamreader"
	"cigarmender/log"
	"cigarmender/reference"
	"os"
	"runtime"
	"runtime/pprof"

	"log/slog"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Verbose("Task completed", "task", name, "elapsed", elapsed)
}

func main() {
	defer timeTrack(time.Now(), "CIGARMender execution")

	log.SetupLogger(false)
	args := args.ParseArgs()
	log.SetupLogger(args.Verbose)

	slog.Info("Running cigarmender", "args", args)
	index := buildIndex(args)
	slog.Info(index.Summary())

	bamVisitor := getVisitor(args)

	// CPU profiling
	count, err := captureCPU(func() (int, error) {
		return bamreader.ReadBam(args.Input, bamVisitor, index, args)
	})
	captureAllocs()

	if err != nil {
		log.Fatalf("Error reading bam %v", err)
	}
	slog.Info("Processed reads", "readCount", count)

	slog.Info(bamVisitor.Summary())
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
		slog.Info("Performing dry run, will read bam and report deletion statistics")
		return bamreader.NewDelCounter()
	} else {
		log.Verbose("Not dry run, centring deletions")
		slog.Info("Rewriting bam file", "input", args.Input, "output", args.OutputDir)
		return bamreader.NewDelCentrer()
	}
}

func captureCPU(operation func() (int, error)) (int, error) {
	f, err := os.Create("cpu.pb.gz")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		return 0, err
	}
	defer pprof.StopCPUProfile()
	i, err := operation()
	return i, err
}
func captureHeap() error {
	runtime.GC() // run GC first to capture only the most current objects
	f, err := os.Create("heap.pb.gz")
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.Lookup("heap").WriteTo(f, 0)
}
func captureAllocs() error {
	f, err := os.Create("allocs.pb.gz")
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.Lookup("allocs").WriteTo(f, 0)
}
