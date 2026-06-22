package bamreader

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"

	"cigarmender/args"
)

type ReadVisitor interface {
	Visit(read *sam.Record, refName string) error
	Summary() string
}

/**
* ReadBam reads in the bam file and applies the visitor function to each read.
* Returns a count of the number of reads processed
 */
func ReadBam(bamfileStr string, visitor ReadVisitor, args args.Args) (int, error) {
	bamf, err := os.Open(bamfileStr)
	if err != nil {
		log.Fatalf("Could not open file : %s - %v", bamfileStr, err)
	}
	defer bamf.Close()

	bamreader, err := bam.NewReader(bamf, args.Threads)
	if err != nil {
		log.Fatalf("Could not create reader for %s: %v", bamfileStr, err)
	}

	iter, err := bam.NewIterator(bamreader, nil)
	if err != nil {
		log.Fatalf("Could not create iterator %v", err)
	}

	readCount := 0

	for iter.Next() {
		read := iter.Record()
		if err = visitor.Visit(read, ""); err != nil {
			return readCount, fmt.Errorf("Error processing read : %s", read)
		}

		readCount++
		if readCount%200000 == 0 {
			log.Printf("Processed %d reads", readCount)
		}
	}

	if err := iter.Error(); err != nil {
		if err == io.EOF {
			fmt.Printf("Successfully processed %s to end of the BAM file.", bamfileStr)
		} else {
			return readCount, fmt.Errorf("Error occurred reading %s: %v", bamfileStr, err)
		}
	}
	return readCount, nil
}
