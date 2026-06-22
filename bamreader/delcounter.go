package bamreader

import (
	"cigarmender/reference"
	"fmt"

	"github.com/biogo/hts/sam"
)

func NewDelCounter(index reference.HPIndex) *DelCounter {
	return &DelCounter{hpIndex: index}
}

type DelCounter struct {
	ReadWithDel   int
	InHomopolymer int
	Count         int
	Len           int
	hpIndex       reference.HPIndex
}

func (d *DelCounter) Summary() string {
	return fmt.Sprintf("Counted %d unique reads and %d total deletions with average length %.2f\n%d deletions in homopolymers",
		d.ReadWithDel, d.Count, float64(d.Len)/float64(d.Count), d.InHomopolymer)
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCounter) Visit(read *sam.Record, s string) error {

	hasDel := false
	rpos := 0

	for _, cigarop := range read.Cigar {
		rpos += cigarop.Type().Consumes().Reference
		if cigarop.Type() == sam.CigarDeletion {
			d.Count++
			d.Len += cigarop.Len()
			hasDel = true
			_, found := d.hpIndex.Search(read.Ref.Name(), reference.NewRange(rpos, rpos+cigarop.Len()))
			if found {
				d.InHomopolymer++
			}
		}
	}
	if hasDel {
		d.ReadWithDel++
	}
	return nil
}
