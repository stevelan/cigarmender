package main

import (
	"cigarmender/args"
	"cigarmender/bamreader"
	"log"

	"github.com/biogo/hts/sam"
)

func main() {

	args := args.ParseArgs()

	log.Printf("Processing cigarmender with %s", args.ToString())

	delCounter := bamreader.DelCounter{}
	count, err := bamreader.ReadBam(args.Input, &delCounter, args)
	if err != nil {
		log.Fatalf("Error reading bam %v", err)
	}
	log.Printf("Processed %d reads", count)

	log.Printf(delCounter.Summary())
}

func getVisitor(args args.Args) bamreader.ReadVisitor {
	if args.DryRun {
		return &bamreader.DelCounter{}
	} else {
		log.Fatalf("Implement me")
		return nil
	}
}

func VisitRead(r *sam.Record, val string) error {
	log.Printf("%d - %s", r.Pos, r.Cigar)
	return nil
}
