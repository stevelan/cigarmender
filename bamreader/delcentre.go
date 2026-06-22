package bamreader

import (
	"fmt"
	"log"

	"github.com/biogo/hts/sam"
)

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
func (d *DelCenterer) Visit(read *sam.Record, s string) error {
	qpos := 0
	rpos := 0

	for _, cigarop := range read.Cigar {
		if cigarop.Type() == sam.CigarDeletion { // deletion doesn't advance query
			// check if hp

		} else {
			qpos += cigarop.Type().Consumes().Reference
			rpos += cigarop.Type().Consumes().Query
		}
	}
	log.Fatalf("Implement me - delcentre.go")
	return nil
}
