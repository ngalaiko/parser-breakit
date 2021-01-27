package main

import (
	"context"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"

	parser "github.com/ngalaiko/parser-breakit"
)

func main() {
	depth := flag.Int64("d", 0, `Parsing recursion depth. For example, if set to 1, all pages that are
linked from a found page wil be also parsed`)
	concurrency := flag.Int64("p", 1, "How many pages to parse concurrently")
	verbose := flag.Bool("v", false, "Verbose logging")
	out := flag.String("o", "-", "Output filename")

	flag.Parse()

	if *concurrency < 1 {
		log.Fatalf("concurrency should be set to at least 1")
	}

	writer, err := getWriter(*out)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ctx := context.Background()

	aa, err := parser.New(*verbose).Parse(ctx, *depth, *concurrency)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	output(aa, writer)
}

func output(aa []*parser.Article, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)

	if err := csvWriter.Write([]string{"Link", "Published", "Title", "Preamble", "First Paragraph"}); err != nil {
		return err
	}

	for _, a := range aa {
		columns := []string{a.URL.String(), a.PublishedAt.String(), a.Title, a.Preamble}
		if a.Summary != nil {
			columns = append(columns, *a.Summary)
		} else {
			columns = append(columns, "")
		}

		if err := csvWriter.Write(columns); err != nil {
			return err
		}
	}
	return nil
}

func getWriter(filename string) (io.Writer, error) {
	writer := os.Stdout
	if filename == "-" {
		return writer, nil
	}
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}
