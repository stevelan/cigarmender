package bamio

import (
	"fmt"
	"strings"

	"github.com/biogo/hts/sam"
)

// NewModCounter create a ModCounter for counting the number of m6a mods on a read
func NewModCounter() *ModCounter {
	return &ModCounter{
		Modified: make([]int, 11),
	}
}

// ModCounter is a BAM visitor that counts modifications in reads
type ModCounter struct {
	Reads      int
	Unmodified int
	Modified   []int
}

func (d *ModCounter) Summary() string {
	return fmt.Sprintf("Processed %d reads, found %d with m6A and %d without any", d.Reads, d.Modified, d.Unmodified)
}

const ML_THRESHOLD = uint8(230) // minimum ML value to consider a base as modified, 255 * 0.9 = 230
/**
* Counts the number of reads with modifications
 */
func (d *ModCounter) Process(read *sam.Record, _ *BamWriter) error {

	d.Reads++
	M6ACount, err := HasM6AAboveThreshold(read, ML_THRESHOLD)
	if err != nil {
		return fmt.Errorf("processing read as unmodified - %v", err)
	}
	if M6ACount == 0 {
		d.Unmodified++
	} else if M6ACount > 10 {
		d.Modified[10]++
	} else {
		d.Modified[M6ACount]++
	}
	return nil
}

func HasM6AAboveThreshold(rec *sam.Record, minML uint8) (int, error) {
	mmAux, ok := rec.Tag([]byte("MM"))
	if !ok {
		return 0, nil
	}

	mm, ok := mmAux.Value().(string)
	if !ok {
		return 0, fmt.Errorf("MM tag has unexpected type %T", mmAux.Value())
	}

	mlAux, ok := rec.Tag([]byte("ML"))
	if !ok {
		return 0, fmt.Errorf("MM tag present but ML tag missing")
	}

	ml, err := auxBytes(mlAux)
	if err != nil {
		return 0, err
	}

	mlIdx := 0

	count := 0
	for _, group := range strings.Split(mm, ";") {
		if group == "" {
			continue
		}

		head, nums, ok := strings.Cut(group, ",")
		if !ok {
			continue
		}

		nCalls := countMMDeltas(nums)

		isM6A := strings.HasPrefix(head, "A+a")

		for i := 0; i < nCalls; i++ {
			if mlIdx >= len(ml) {
				return 0, fmt.Errorf("ML shorter than MM calls")
			}

			if isM6A && ml[mlIdx] >= minML {
				count++
			}

			mlIdx++
		}
	}

	return count, nil
}

func countMMDeltas(nums string) int {
	if nums == "" {
		return 0
	}

	n := 0
	for _, part := range strings.Split(nums, ",") {
		if part != "" {
			n++
		}
	}
	return n
}

func auxBytes(a sam.Aux) ([]byte, error) {
	v, ok := a.Value().([]byte)
	if !ok {
		return nil, fmt.Errorf("unsupported ML value type %T", a.Value())
	}
	return v, nil
}
