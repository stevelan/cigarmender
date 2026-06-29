package reference

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sort"

	"strings"

	"github.com/stevelan/cigarmender/internal/log"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq/linear"
)

// Range - a genomic range with a Start and end position
type Range struct {
	Start int
	End   int
}

// NewRange creates a new range with the largest value guaranteed to be after the smallest
func NewRange(s, e int) Range {
	if e <= s {
		return Range{Start: e, End: s}
	}
	return Range{Start: s, End: e}
}

// IsWithin returns true if the target range is within the receiver range
func (r *Range) IsWithin(target Range) bool {
	return r.Start <= target.Start && target.End < r.End
}

func (r *Range) Len() int {
	return r.End - r.Start
}

// String receiver function for Range
func (r *Range) String() string {
	return fmt.Sprintf("Range{%d to %d}", r.Start, r.End)
}

/**
* Sort ranges by the Start position within the range
 */
func sortRanges(ranges []Range) {
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
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
		if e.Start < t.Start {
			return -1
		}

		return 1
	})
	if found {
		return ref.index[id][idx], true
	}
	return Range{}, false
}

// IndexHomopolymers - Scan through the reference genome and collect homopolymers in an index. Homopolymers are represented
// as a range with a start and end position within the sequence
func IndexHomopolymers(refFastaPath string, hpMinSize int, bases []string) (homopolymerIndex *RefIndex, err error) {
	slog.Info("Building homopolymer index", "reference", refFastaPath, "min-hp-size", hpMinSize, "bases", bases)
	reference, err := os.Open(refFastaPath)
	if err != nil {
		return nil, fmt.Errorf("error opening reference file %s - %v", refFastaPath, err)
	}
	defer log.CloseAndLog("closing readonly reference file", reference.Close)

	index := make(map[string][]Range)

	template := linear.NewSeq("", nil, alphabet.DNA)
	reader := fasta.NewReader(reference, template)
	scanner := seqio.NewScanner(reader)

	// Create a slice of booleans to quickly check whether a base should be indexed or not
	basesToIndex := make(map[alphabet.Letter]bool)
	for _, b := range bases {
		if len(b) != 1 {
			return nil, fmt.Errorf("base must be a single character: %q", b)
		}
		basesToIndex[alphabet.Letter(strings.ToUpper(b)[0])] = true
	}

	for scanner.Next() {
		sequence := scanner.Seq().(*linear.Seq)
		index[sequence.ID] = findHomopolymersInSeq(sequence, hpMinSize, basesToIndex)
		slog.Debug("Indexed reference sequence", "name", sequence.Name(), "hpCount", len(index[sequence.ID]))
	}

	if err = scanner.Error(); err != nil {
		return nil, fmt.Errorf("error scanning reference %v", err)
	}
	return newRefIndex(index), nil
}

func findHomopolymersInSeq(sequence *linear.Seq, hpMinSize int, shouldIndex map[alphabet.Letter]bool) []Range {
	foundHps := make([]Range, 0)
	var lastBase alphabet.Letter = 0
	inHp := false
	newHpRange := Range{}

	for baseIdx := 0; baseIdx < sequence.Len(); baseIdx++ {
		base := sequence.At(baseIdx).L

		if base == lastBase {
			// this base same as last base Start a new homopolymer range or continue
			if !inHp {
				inHp = true
				newHpRange.Start = baseIdx - 1 // use previous Start index as Start position
			}
			continue
		} else if inHp && (baseIdx-newHpRange.Start) >= hpMinSize {
			// this base different to last base and we are in a homopolymer and it is the minimum size
			// store range
			newHpRange.End = baseIdx
			if shouldIndex[lastBase] {
				foundHps = append(foundHps, newHpRange)
			}
		}
		// reset variables
		lastBase = base
		inHp = false
		newHpRange = Range{}

	}
	// grab homopolymer at end of the sequence
	if inHp && (sequence.Len()-newHpRange.Start) >= hpMinSize {
		newHpRange.End = sequence.Len()
		if shouldIndex[lastBase] {
			foundHps = append(foundHps, newHpRange)
		}
	}

	return foundHps
}
