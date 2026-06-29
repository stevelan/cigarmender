package bamreader

import (
	"cigarmender/log"
	"fmt"
	"os"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

// NewBamWriter opens a file for writing with the given bam header.
// Caller must call Close() on the returned writer
func NewBamWriter(outputPath string, bamHeader *sam.Header, compression int, threads int) (*BamWriter, error) {

	out, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create output BAM: %w", err)
	}

	var writer *bam.Writer

	log.Verbose("Compression settings", "level", compression, "concurrency", threads)
	writer, err = bam.NewWriterLevel(out, bamHeader, compression, threads)

	if err != nil {
		return nil, fmt.Errorf("create BAM writer: %w", err)
	}

	log.Verbose("BAM Writer created", "path", outputPath)

	return &BamWriter{
		path:   outputPath,
		writer: writer,
	}, nil

}

type BamWriter struct {
	path   string
	writer *bam.Writer
}

func (bw *BamWriter) Close() {
	bw.writer.Close()
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
