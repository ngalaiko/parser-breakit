package main

import (
	"context"
	"flag"
	"log"

	parser "github.com/ngalaiko/parser-breakit"
)

func main() {
	depth := flag.Int64("d", 0, `Parsing recursion depth. For example, if set to 1, all pages that are
linked from a found page wil be also parsed`)
	concurrency := flag.Int64("p", 1, "How many pages to parse concurrently")

	flag.Parse()

	if *concurrency < 1 {
		log.Fatalf("concurrency should be set to at least 1")
	}

	ctx := context.Background()

	aa, err := parser.New().Parse(ctx, *depth, *concurrency)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	// TODO: prettify the output
	log.Print(aa)
}
