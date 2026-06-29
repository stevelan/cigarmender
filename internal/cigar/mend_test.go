package cigar

import (
	"testing"

	"github.com/stevelan/cigarmender/internal/reference"

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
	// homopolymer run starting at the deletion and continuing for 10 bases
	output, modified := rewriteCigar(priorMatch, deletion, nextMatch, reference.NewRange(10, 20), 10)
	if !modified {
		testing.Fatal("Expected modified")
	}

	totalLen := output[0].Len() + output[1].Len() + output[2].Len()
	if totalLen != 25 {
		testing.Fatalf("Expected total length of 25 - %d - %v", totalLen, output)
	}
	// centred, 2 bases at start of run and 3 bases at the end
	if output[0].Len() != 12 {
		testing.Fatalf("Expected len 12 - %d - %v", output[0], output)
	}
	if output[1].Len() != 5 {
		testing.Fatalf("Expected len 5 - %d - %v", output[1], output)
	}
	if output[2].Len() != 8 {
		testing.Fatalf("Expected len 8 - %d - %v", output[2], output)
	}
}
