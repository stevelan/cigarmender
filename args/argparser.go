package args

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"
)

type Args struct {
	Input     string // input BAM file to mend
	SampleCSV string // optional two column csv file with SampleName and Sample-BAM-file of the bams to be processed
	OutputDir string // output directory
	Threads   int    // number of threads to use
	DryRun    bool
	Bases     []string // list of bases to mend, defaults to A,C,G,T,U
	LogFile   string   // log file to write to
}

func (a Args) ToString() string {
	return fmt.Sprintf("Args: %#v", a)
}

func ParseArgs() Args {
	args := Args{}

	flag.StringVar(&args.Input, "input", "", "required: input BAM file")
	flag.StringVar(&args.OutputDir, "output", "", "required: output BAM file")
	flag.IntVar(&args.Threads, "threads", 4, "optional: number of threads")
	flag.BoolVar(&args.DryRun, "dry-run", false, "optional: print changes without writing output")

	bases := flag.String("bases", "A,C,G,T,U", "optional: set of bases to check for homopolymer runs")
	args.Bases = strings.Split(*bases, ",")
	flag.Parse()
	if err := validate(&args); err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	return args
}

func validate(arg *Args) error {
	if arg.Input == "" {
		return fmt.Errorf("Input is required but was blank")
	}

	supportedBases := []string{"A", "C", "G", "T", "U"}
	for _, base := range arg.Bases {
		if !slices.Contains(supportedBases, base) {
			return fmt.Errorf("Unsupported base {%s} in bases {%s}", base, arg.Bases)
		}
	}
	return nil
}
