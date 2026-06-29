package bamio

import (
	"fmt"
	"log/slog"

	"github.com/stevelan/cigarmender/internal/cigar"
	"github.com/stevelan/cigarmender/internal/reference"

	"github.com/biogo/hts/sam"
)

// NewDelCentrer creates a DelCenterer
func NewDelCentrer() *DelCentrer {
	return &DelCentrer{}
}

type DelCentrer struct {
	Deletions    int
	Homopolymers int
	Rewrites     int
}

func (d DelCentrer) Summary() string {
	return fmt.Sprintf("Found %d homopolymers out of %d deletions, with %d rewrites", d.Homopolymers, d.Deletions, d.Rewrites)
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCentrer) Process(read *sam.Record, hpIndex *reference.RefIndex, bamWriter *BamWriter) error {

	newCigar, stats := cigar.ProcessCigar(read, hpIndex)

	d.Deletions += stats.Deletions
	d.Homopolymers += stats.Homopolymers
	d.Rewrites += stats.Rewrites

	if stats.Modified {
		slog.Debug("Writing new CIGAR for read", "read", read.Name, "cigar", newCigar)
		return bamWriter.WriteToBam(read, newCigar)
	} else {
		slog.Debug("Writing existing unmodified read", "read", read.Name, "cigar", newCigar)
		return bamWriter.WriteToBamExisting(read)
	}
}
