package reference

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sort"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq/linear"
)

/**
* Range - a genomic range with a start and end position
 */
type Range struct {
	start int
	end   int
}

func NewRange(s, e int) Range {
	if e <= s {
		return Range{start: e, end: s}
	}
	return Range{start: s, end: e}
}

func (r *Range) IsWithin(loc int) bool {
	return r.start <= loc && loc < r.end
}

func (r *Range) ToString() string {
	return fmt.Sprintf("Range{%d to %d}", r.start, r.end)
}

/**
* Sort ranges by the start position within the range
 */
func sortRanges(ranges []Range) {
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start < ranges[j].start
	})
}

type HPIndex struct {
	index map[string][]Range
}

/**
* IndexHomopolymers - Scan through the reference genome and collect homopolymers in an index.
 */
func IndexHomopolymers(refFastaPath string, hpMinSize int, bases []string) (*HPIndex, error) {
	reference, err := os.Open(refFastaPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening reference file %s - %v", refFastaPath, err)
	}
	defer reference.Close()

	index := make(map[string][]Range)

	template := linear.NewSeq("", nil, alphabet.DNA)
	reader := fasta.NewReader(reference, template)
	scanner := seqio.NewScanner(reader)

	toIndex := make(map[alphabet.Letter]bool)
	for _, b := range bases {
		toIndex[alphabet.Letter(b[0])] = true
	}

	for scanner.Next() {
		sequence := scanner.Seq().(*linear.Seq)

		index[sequence.ID] = make([]Range, 0)

		var lastBase alphabet.Letter = 0
		inHp := false
		newRange := Range{}

		for baseIdx := 0; baseIdx < sequence.Len(); baseIdx++ {
			base := sequence.At(baseIdx)
			if base.L == lastBase {
				if !inHp {
					inHp = true
					newRange.start = baseIdx - 1 // use previous start index as start position
				}
				continue
			} else if inHp && (baseIdx-newRange.start) >= hpMinSize {
				// store range
				newRange.end = baseIdx
				if _, ok := toIndex[lastBase]; ok {
					index[sequence.ID] = append(index[sequence.ID], newRange)
				}
			}
			// reset variables
			lastBase = base.L
			inHp = false
			newRange = Range{}

		}
		// grab homopolymer at end of the sequence
		if inHp && (sequence.Len()-newRange.start) >= hpMinSize {
			newRange.end = sequence.Len()
			if _, ok := toIndex[lastBase]; ok {
				index[sequence.ID] = append(index[sequence.ID], newRange)
			}
		}
	}

	if err = scanner.Error(); err != nil {
		return nil, fmt.Errorf("Error scanning reference %v", err)
	}

	// sort index
	for _, ranges := range index {
		sortRanges(ranges)
	}

	return &HPIndex{index: index}, nil
}

func (hpIndex *HPIndex) Search(id string, query Range) (int, bool) {
	if hpIndex.index[id] == nil {
		log.Fatalf("Unknown ID in index, was the alignment done with the same reference as the index : %s", id)
	}
	return slices.BinarySearchFunc(hpIndex.index[id], query, func(e Range, t Range) int {
		if e.IsWithin(t.start) {
			return 0
		}
		if e.start < t.start {
			return -1
		}

		return 1
	})
}
