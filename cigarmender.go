package main

import (
	"cigarmender/args"
	"cigarmender/bamreader"
	"log"

	"time"

	"github.com/biogo/hts/sam"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func main() {
	defer timeTrack(time.Now(), "CIGARMender execution")

	args := args.ParseArgs()

	log.Printf("Processing cigarmender with %s", args.ToString())

	delCounter := getVisitor(args)
	count, err := bamreader.ReadBam(args.Input, delCounter, args)
	if err != nil {
		log.Fatalf("Error reading bam %v", err)
	}
	log.Printf("Processed %d reads", count)

	log.Println(delCounter.Summary())
}

func getVisitor(args args.Args) bamreader.ReadVisitor {
	if args.DryRun {
		return &bamreader.DelCounter{}
	}

	log.Fatalf("Implement me")
	return nil
}

func VisitRead(r *sam.Record, val string) error {
	log.Printf("%d - %s", r.Pos, r.Cigar)
	return nil
}
