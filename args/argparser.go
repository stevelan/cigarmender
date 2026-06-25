package args

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strings"
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
}

func (a Args) String() string {
	return fmt.Sprintf("Args: %#v", a)
}

func ParseArgs() Args {
	args := Args{}

	flag.StringVar(&args.Input, "input", "", "required: input BAM file")
	flag.StringVar(&args.OutputDir, "output", "", "required: output BAM file")
	flag.StringVar(&args.Reference, "reference", "", "required: reference that the alignment was performed against")
	flag.IntVar(&args.Threads, "threads", defaultThreads(), "optional: number of threads, defaults to num CPUS minus two")
	flag.BoolVar(&args.DryRun, "dry-run", false, "optional: print changes without writing output")
	flag.IntVar(&args.HomopolymerSize, "min-hp", 3, "optional: number of repeat bases to be considered a homopolymer")
	flag.BoolVar(&args.Verbose, "verbose", false, "optional: enables verbose logging")
	flag.IntVar(&args.CompressionLevel, "compress-level", 3, "optional: Changes the compression level between 1 (best speed) and 9 (best compression)")

	bases := flag.String("bases", "A,C,G,T,U", "optional: set of bases to check for homopolymer runs")
	args.Bases = strings.Split(*bases, ",")
	flag.Parse()
	if err := validate(&args); err != nil {
		flag.Usage()
		slog.Error("Usage error", "err", err)
		os.Exit(1)
	}

	return args
}

func defaultThreads() int {
	return max(runtime.NumCPU()-2, 1)
}

func validate(arg *Args) error {
	if arg.Input == "" {
		return fmt.Errorf("Input is required but was blank")
	}

	if arg.OutputDir == "" {
		return fmt.Errorf("Output directory is required but was blank")
	}

	if arg.Reference == "" {
		return fmt.Errorf("Reference is required by but was blank")
	}

	if arg.CompressionLevel < 1 || arg.CompressionLevel > 9 {
		return fmt.Errorf("Compression level invalid, valid values between 1 and 9, but got: %d", arg.CompressionLevel)
	}

	supportedBases := []string{"A", "C", "G", "T", "U"}
	for _, base := range arg.Bases {
		if !slices.Contains(supportedBases, base) {
			return fmt.Errorf("Unsupported base {%s} in bases {%s}", base, arg.Bases)
		}
	}
	return nil
}
