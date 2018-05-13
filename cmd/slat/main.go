package main

import (
	"flag"
	"log"
	"os"

	"github.com/pteichman/slat"
)

var outdir = flag.String("o", ".", "output directory")

func main() {
	flag.Parse()

	if err := os.MkdirAll(*outdir, os.ModePerm); err != nil {
		log.Fatalf("ERROR: creating %s: %s", *outdir, err)
	}

	if flag.NArg() > 0 {
		archive := flag.Arg(0)
		if err := slat.ExportArchiveFile(*outdir, archive); err != nil {
			log.Fatalf("ERROR: exporting %s: %s", archive, err)
		}
		os.Exit(0)
	}

	if token := os.Getenv("SLACK_API_TOKEN"); token != "" {
		if err := slat.ExportHistory(*outdir, token); err != nil {
			log.Fatalf("ERROR: catching up history: %s", err)
		}
		os.Exit(0)
	}

	log.Fatalf("ERROR: no archive or API token provided")
}
