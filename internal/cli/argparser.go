package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/stevelan/cigarmender/internal/log"
)

// Args struct to hold the command line args
type Args struct {
	Input string // input BAM file to mend
	// SampleCSV       string // optional two column csv file with SampleName and Sample-BAM-file of the bams to be processed
	OutputDir string // output directory
	Threads   int    // number of threads to use
	DryRun    bool
	Bases     []string // list of bases to mend, defaults to A,C,G,T,U
	// LogFile         string   // log file to write to
	Reference        string // genome reference that the alignment was run against
	HomopolymerSize  int    // minimum length required to be considered a homopolymer
	Verbose          bool
	CompressionLevel int
	Command          string
}

func (a Args) String() string {
	return fmt.Sprintf("Args: %#v", a)
}

func ParseArgs() Args {
	// default args for optional params
	args := Args{
		OutputDir:        "ash",
		Threads:          defaultThreads(),
		DryRun:           false,
		HomopolymerSize:  4,
		Verbose:          false,
		CompressionLevel: 3,
	}

	// required args
	flag.StringVar(&args.Input, "input", "", "required: input BAM file")
	flag.StringVar(&args.Reference, "ref", "", "required: reference that the BAMs were aligned against")

	// optional args
	flag.StringVar(&args.OutputDir, "output", args.OutputDir, "optional: output directory")
	flag.IntVar(&args.Threads, "threads", args.Threads, "optional: number of threads, defaults to num CPUs minus two")
	flag.BoolVar(&args.DryRun, "dry-run", args.DryRun, "optional: print changes without writing output")
	flag.IntVar(&args.HomopolymerSize, "min-hp", args.HomopolymerSize, "optional: number of repeat bases to be considered a homopolymer")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "optional: enables verbose logging")
	flag.IntVar(&args.CompressionLevel, "compress-level", args.CompressionLevel, "optional: Changes the compression level between 1 (best speed) and 9 (best compression)")

	flag.StringVar(&args.Command, "command", "cigarmender", "optional: Alternate commands for this cli. This field is mostly used for debugging")

	bases := flag.String("bases", "A,C,G,T,U", "optional: set of bases to check for homopolymer runs")
	args.Bases = strings.Split(*bases, ",")

	flag.Parse()
	if err := validate(&args); err != nil {
		flag.Usage()
		log.Error("Usage error", "err", err)
		os.Exit(1)
	}

	return args
}

func defaultThreads() int {
	return max(runtime.NumCPU()-2, 1)
}

func validate(arg *Args) error {
	if arg.Input == "" {
		return fmt.Errorf("input is required but was blank")
	}

	if arg.OutputDir == "" {
		return fmt.Errorf("output directory is required but was blank")
	}

	if arg.Reference == "" {
		return fmt.Errorf("reference is required by but was blank")
	}

	validCommands := map[string]bool{
		"cigarmender": true,
		"readfilter":  true,
	}
	if !validCommands[arg.Command] {
		return fmt.Errorf("command was not valid, got %s", arg.Command)
	}

	if arg.CompressionLevel < 1 || arg.CompressionLevel > 9 {
		return fmt.Errorf("compression level invalid, valid values between 1 and 9, but got: %d", arg.CompressionLevel)
	}

	supportedBases := []string{"A", "C", "G", "T", "U"}
	for _, base := range arg.Bases {
		if !slices.Contains(supportedBases, base) {
			return fmt.Errorf("unsupported base {%s} in bases {%s}", base, arg.Bases)
		}
	}
	return nil
}
