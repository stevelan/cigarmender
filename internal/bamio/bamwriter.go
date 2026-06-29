package bamio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stevelan/cigarmender/internal/cli"
	"github.com/stevelan/cigarmender/internal/log"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

// NewBamWriter opens a file for writing with the given bam header.
// Caller must call Close() on the returned writer
func NewBamWriter(inputFile string, bamHeader *sam.Header, args cli.Args) (*BamWriter, error) {

	outputPath := getOutputFile(inputFile, args)
	out, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create output BAM: %w", err)
	}

	var writer *bam.Writer

	log.Verbose("Compression settings", "level", args.CompressionLevel, "concurrency", args.Threads)
	writer, err = bam.NewWriterLevel(out, bamHeader, args.CompressionLevel, args.Threads)

	if err != nil {
		return nil, fmt.Errorf("create BAM writer: %w", err)
	}

	log.Verbose("BAM Writer created", "path", outputPath)

	return &BamWriter{
		path:   outputPath,
		writer: writer,
	}, nil

}

func getOutputFile(bamfileStr string, args cli.Args) string {
	baseInputFile := filepath.Base(bamfileStr)
	extension := filepath.Ext(bamfileStr)
	baseOutFile := strings.TrimSuffix(baseInputFile, extension) + ".mended" + extension

	log.Verbose("Writing to output file", "file", baseOutFile, "outputDir", args.OutputDir)

	return filepath.Join(args.OutputDir, baseOutFile)
}

type BamWriter struct {
	path   string
	writer *bam.Writer
}

func (bw *BamWriter) Close() error {
	return bw.writer.Close()
}

func (bw *BamWriter) String() string {
	return bw.path
}

func (bw *BamWriter) WriteToBam(record *sam.Record, newCigar sam.Cigar) error {

	// check that the cigar is a valid length for the sequence
	if !record.Cigar.IsValid(record.Seq.Length) {
		return fmt.Errorf("invalid CIGAR for %s: %s", record.Name, record.Cigar)
	}

	record.Cigar = newCigar
	// these tags are stale after a change to a cigar string
	// can be regenerated with samtools calmd
	removeAuxTags(record, "MD", "NM")

	return bw.WriteToBamExisting(record)
}

func (bw *BamWriter) WriteToBamExisting(record *sam.Record) error {
	if err := bw.writer.Write(record); err != nil {
		return fmt.Errorf("write BAM record %s: %w", record.Name, err)
	}
	return nil
}

func removeAuxTags(rec *sam.Record, tags ...string) {
	remove := make(map[sam.Tag]bool, len(tags))
	for _, tag := range tags {
		remove[sam.NewTag(tag)] = true
	}

	aux := rec.AuxFields[:0]
	for _, field := range rec.AuxFields {
		if !remove[field.Tag()] {
			aux = append(aux, field)
		}
	}
	rec.AuxFields = aux
}
