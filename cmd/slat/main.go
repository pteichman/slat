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
		log.Printf("Creating %s: %s", *outdir, err)
		os.Exit(1)
	}

	if flag.NArg() > 0 {
		archive := flag.Arg(0)
		if err := slat.ExportArchiveFile(*outdir, archive); err != nil {
			log.Printf("Exporting %s: %s", archive, err)
		}
	}
}
