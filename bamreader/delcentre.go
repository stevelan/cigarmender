package bamreader

import (
	"fmt"

	"cigarmender/reference"

	"github.com/biogo/hts/sam"
)

// NewDeletionCentrer creates a DelCenterer
func NewDeletionCentrer() DelCenterer {
	return DelCenterer{}
}

type DelCenterer struct {
	DelCount int
	HPCount  int
}

func (d *DelCenterer) Summary() string {
	return fmt.Sprintf("Found %d homopolymers out of %d deletions", d.HPCount, d.DelCount)
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCenterer) Visit(read *sam.Record, hpIndex *reference.RefIndex) error {
	qpos := 0
	rpos := 0

	for _, cigarop := range read.Cigar {
		if cigarop.Type() == sam.CigarDeletion { // deletion doesn't advance query
			// check if hp
			query := reference.NewRange(rpos, rpos+cigarop.Len())
			hp, found := hpIndex.Search(read.Ref.Name(), query)
			if found {
				println("Found homopolymer for read : %s", hp.String())
			}

		} else {
			rpos += cigarop.Type().Consumes().Reference
			qpos += cigarop.Type().Consumes().Query
		}
	}
	// log.Fatalf("Implement me - delcentre.go")
	return nil
}
