package main

import (
	"cigarmender/args"
	"cigarmender/bamreader"
	"cigarmender/reference"
	"fmt"
	"log"

	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s completed, took %s", name, elapsed)
}

func main() {
	defer timeTrack(time.Now(), "CIGARMender execution")

	args := args.ParseArgs()

	log.Printf("Processing cigarmender with %s", args.ToString())
	index := buildIndex(args)
	fmt.Println(index.Summary())

	delCounter := getVisitor(args, index)
	count, err := bamreader.ReadBam(args.Input, delCounter, args)
	if err != nil {
		log.Fatalf("Error reading bam %v", err)
	}
	log.Printf("Processed %d reads", count)

	log.Println(delCounter.Summary())
}

func buildIndex(args args.Args) reference.HPIndex {
	defer timeTrack(time.Now(), "Building index")
	hpindex, err := reference.IndexHomopolymers(args.Reference, 3, args.Bases)
	if err != nil {
		log.Fatalf("Could not build index for reference %s - %v", args.Reference, err)
	}
	return *hpindex
}

func getVisitor(args args.Args, index reference.HPIndex) bamreader.ReadVisitor {
	if args.DryRun {
		return bamreader.NewDelCounter(index)
	}

	log.Fatalf("Implement me")
	return nil
}
