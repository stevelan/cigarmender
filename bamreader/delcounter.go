package bamreader

import (
	"cigarmender/reference"
	"fmt"

	"github.com/biogo/hts/sam"
)

// NewDelCounter create a DelCounter with an index as a private field.
func NewDelCounter() *DelCounter {
	return &DelCounter{}
}

// DelCounter is a BAM visitor that counts deletions in reads. It's used when doing a dry run.
type DelCounter struct {
	ReadWithDel   int
	InHomopolymer int
	Count         int
	Len           int
}

func (d DelCounter) Summary() string {
	return fmt.Sprintf("Counted %d unique reads and %d total deletions with average length %.2f\n%d deletions in homopolymers",
		d.ReadWithDel, d.Count, float64(d.Len)/float64(d.Count), d.InHomopolymer)
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCounter) Visit(read *sam.Record, hpIndex *reference.RefIndex, _ *BamWriter) error {

	hasDel := false
	rpos := 0

	for _, cigarop := range read.Cigar {
		rpos += cigarop.Type().Consumes().Reference
		if cigarop.Type() == sam.CigarDeletion {
			d.Count++
			d.Len += cigarop.Len()
			hasDel = true
			_, found := hpIndex.Search(read.Ref.Name(), reference.NewRange(rpos, rpos+cigarop.Len()))
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
