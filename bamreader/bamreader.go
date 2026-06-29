package bamreader

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"cigarmender/args"
	"cigarmender/log"
	"cigarmender/reference"
	"path/filepath"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

type ReadVisitor interface {
	Visit(read *sam.Record, hpIndex *reference.RefIndex, bamWriter *BamWriter) error
	Summary() string
}

// log progress each N reads
const progress = 250000

// ReadBam reads in the bam file and applies the visitor function to each read.
// Returns a count of the number of reads processed
func ReadBam(bamfileStr string, visitor ReadVisitor, hpIndex *reference.RefIndex, args args.Args) (int, error) {
	bamf, err := os.Open(bamfileStr)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not open file : %s - %v", bamfileStr, err)
	}
	defer bamf.Close()

	bamreader, err := bam.NewReader(bamf, args.Threads)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not create reader for %s: %v", bamfileStr, err)
	}

	iter, err := bam.NewIterator(bamreader, nil)
	if err != nil {
		return 0, fmt.Errorf("ReadBam could not create iterator %v", err)
	}
	outputFile := getOutputFile(bamfileStr, args)

	bamWriter, err := NewBamWriter(outputFile, bamreader.Header(), args.CompressionLevel, args.Threads)
	if err != nil {
		return 0, fmt.Errorf("Creating bam writer - %v", err)
	}
	defer bamWriter.Close()

	readCount := 0

	for iter.Next() {
		read := iter.Record()
		if err = visitor.Visit(read, hpIndex, bamWriter); err != nil {
			return readCount, fmt.Errorf("Error processing read : %s", read)
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

func getOutputFile(bamfileStr string, args args.Args) string {
	baseInputFile := filepath.Base(bamfileStr)
	extension := filepath.Ext(bamfileStr)
	baseOutFile := strings.TrimSuffix(baseInputFile, extension) + ".mended" + extension
	log.Verbose("Writing to output file", "file", baseOutFile, "outputDir", args.OutputDir)
	return filepath.Join(args.OutputDir, baseOutFile)
}
