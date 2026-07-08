package bamio

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/stevelan/cigarmender/internal/cli"
	"github.com/stevelan/cigarmender/internal/log"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

type ReadProcessor interface {
	Process(read *sam.Record, bamWriter *BamWriter) error
	Summary() string
}

// log progress each N reads
const progress = 250000

// ReadBam reads in the bam file and applies the ReadProcessor function to each read.
// Returns a count of the number of reads processed
func ReadBam(bamfileStr string, rp ReadProcessor, args cli.Args) (readCount int, retErr error) {
	bamf, err := os.Open(bamfileStr)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not open file : %s - %v", bamfileStr, err)
	}
	defer log.CloseAndLog("closing bam reader %v", bamf.Close)

	bamreader, err := bam.NewReader(bamf, args.Threads)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not create reader for %s: %v", bamfileStr, err)
	}

	iter, err := bam.NewIterator(bamreader, nil)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not create iterator %v", err)
	}

	bamWriter, err := NewBamWriter(bamfileStr, bamreader.Header(), args)
	if err != nil {
		return 0, fmt.Errorf("creating bam writer - %v", err)
	}

	defer func() {
		if err := bamWriter.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	for iter.Next() {
		read := iter.Record()
		if err = rp.Process(read, bamWriter); err != nil {
			return readCount, fmt.Errorf("error processing read : %s", read)
		}

		readCount++
		if readCount%progress == 0 {
			slog.Info("ReadBam Progress", "readCount", readCount)
		}
	}

	if err := iter.Error(); err != nil {
		if err == io.EOF {
			log.Verbose("Successfully processed %s to end of the BAM file.", bamfileStr)
		} else {
			return readCount, fmt.Errorf("ReadBam - Error occurred reading %s: %v", bamfileStr, err)
		}
	}
	return readCount, nil
}
