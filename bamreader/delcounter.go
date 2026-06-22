package bamreader

import (
	"fmt"

	"github.com/biogo/hts/sam"
)

type DelCounter struct {
	ReadWithDel int
	Count       int
	Len         int
}

func (d *DelCounter) Summary() string {
	return fmt.Sprintf("Counted %d unique reads and %d total deletions with average length %.2f", d.ReadWithDel, d.Count, float64(d.Len)/float64(d.Count))
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCounter) Visit(read *sam.Record, s string) error {

	hasDel := false
	for _, cigarop := range read.Cigar {
		if cigarop.Type() == sam.CigarDeletion {
			d.Count++
			d.Len += cigarop.Len()
			hasDel = true
		}
	}
	if hasDel {
		d.ReadWithDel++
	}
	return nil
}
