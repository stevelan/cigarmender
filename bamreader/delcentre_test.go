package bamreader

import (
	"cigarmender/reference"
	"fmt"
	"testing"

	"github.com/biogo/hts/sam"
)

func TestRewriteCigar(testing *testing.T) {

	cigar, err := sam.ParseCigar([]byte("10M5D10M"))
	if err != nil {
		testing.Fatalf("Setup cigar %v", err)
	}

	priorMatch := cigar[0]
	nextMatch := cigar[2]
	deletion := cigar[1]
	output, modified := rewriteCigar(priorMatch, deletion, nextMatch, reference.NewRange(10, 20), 10)
	fmt.Println(output, modified)
}
