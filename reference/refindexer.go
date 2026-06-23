package reference

import (
	"fmt"
	"log"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sort"

	"strings"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq/linear"
)

// Range - a genomic range with a start and end position
type Range struct {
	start int
	end   int
}

// NewRange creates a new range with the largest value guaranteed to be after the smallest
func NewRange(s, e int) Range {
	if e <= s {
		return Range{start: e, end: s}
	}
	return Range{start: s, end: e}
}

// IsWithin returns true if the target range is within the receiver range
func (r *Range) IsWithin(target Range) bool {
	return r.start <= target.start && target.end < r.end
}

// String receiver function for Range
func (r *Range) String() string {
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

// RefIndex is an index of all the homopolymers in the reference fasta keyed by fasta ID
type RefIndex struct {
	index map[string][]Range
}

// newRefIndex creates a new homopolymer index which is just sorted lists of ranges suitable for binary search
func newRefIndex(index map[string][]Range) *RefIndex {
	// sort indexes
	for _, ranges := range index {
		sortRanges(ranges)
	}

	return &RefIndex{index: index}
}

// Summary of the index
func (ref *RefIndex) Summary() string {
	total := 0
	max := 0
	for _, v := range ref.index {
		length := len(v)
		total += length
		if length > max {
			max = length
		}
	}
	return fmt.Sprintf("Homopolymer Index with %d regions and %d total and %d max homopolymers", len(ref.index), total, max)
}

// Search performs a search within the indexed ID for homopolymers that encompass the query Range.
// return the range of the homopolymer, found
func (ref *RefIndex) Search(id string, query Range) (Range, bool) {
	if ref.index[id] == nil {
		log.Fatalf("Unknown ID in index, was the alignment done with the same reference as the index : %s - %s", id,
			slices.Collect(maps.Keys(ref.index)))
	}
	idx, found := slices.BinarySearchFunc(ref.index[id], query, func(e Range, t Range) int {
		if e.IsWithin(t) {
			return 0
		}
		if e.start < t.start {
			return -1
		}

		return 1
	})
	if found {
		return ref.index[id][idx], true
	}
	return Range{}, false
}

// IndexHomopolymers - Scan through the reference genome and collect homopolymers in an index.
func IndexHomopolymers(refFastaPath string, hpMinSize int, bases []string) (homopolymerIndex *RefIndex, err error) {
	slog.Info("Building homopolymer index", "reference", refFastaPath, "min-hp-size", hpMinSize, "bases", bases)
	reference, err := os.Open(refFastaPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening reference file %s - %v", refFastaPath, err)
	}
	defer reference.Close()

	index := make(map[string][]Range)

	template := linear.NewSeq("", nil, alphabet.DNA)
	reader := fasta.NewReader(reference, template)
	scanner := seqio.NewScanner(reader)

	// Create a slice of booleans to quickly check whether a base should be indexed or not
	toIndex := make(map[alphabet.Letter]bool)
	for _, b := range bases {
		if len(b) != 1 {
			return nil, fmt.Errorf("base must be a single character: %q", b)
		}
		toIndex[alphabet.Letter(strings.ToUpper(b)[0])] = true
	}

	for scanner.Next() {
		sequence := scanner.Seq().(*linear.Seq)
		index[sequence.ID] = findHomopolymers(sequence, hpMinSize, toIndex)
		slog.Debug("Indexed reference sequence", "name", sequence.Name(), "hpCount", len(index[sequence.ID]))
	}

	if err = scanner.Error(); err != nil {
		return nil, fmt.Errorf("Error scanning reference %v", err)
	}
	return newRefIndex(index), nil
}

func findHomopolymers(sequence *linear.Seq, hpMinSize int, toIndex map[alphabet.Letter]bool) []Range {
	homopolymers := make([]Range, 0)
	var lastBase alphabet.Letter = 0
	inHp := false
	newRange := Range{}

	for baseIdx := 0; baseIdx < sequence.Len(); baseIdx++ {
		base := sequence.At(baseIdx).L

		if base == lastBase {
			// this base same as last base start a new homopolymer range or continue
			if !inHp {
				inHp = true
				newRange.start = baseIdx - 1 // use previous start index as start position
			}
			continue
		} else if inHp && (baseIdx-newRange.start) >= hpMinSize {
			// this base different to last base and we are in a homopolymer and it is the minimum size or
			// store range
			newRange.end = baseIdx
			if toIndex[lastBase] {
				homopolymers = append(homopolymers, newRange)
			}
		}
		// reset variables
		lastBase = base
		inHp = false
		newRange = Range{}

	}
	// grab homopolymer at end of the sequence
	if inHp && (sequence.Len()-newRange.start) >= hpMinSize {
		newRange.end = sequence.Len()
		if toIndex[lastBase] {
			homopolymers = append(homopolymers, newRange)
		}
	}

	return homopolymers
}
