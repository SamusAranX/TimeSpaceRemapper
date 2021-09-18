package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"TimeSpaceRemapper/remapper"
)

const (
	version = "0.2.1"
)

func main() {
	binName := filepath.Base(os.Args[0])
	binNameNoExt := strings.TrimSuffix(binName, filepath.Ext(binName))

	opts := CommandLineOpts{}

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Printf("%s v%s\n", binNameNoExt, version)
		os.Exit(0)
	}

	if strings.TrimSpace(opts.InputDir) == "" || strings.TrimSpace(opts.OutputDir) == "" {
		fmt.Println("the -i/--input-dir and -o/--output-dir flags are required")
		os.Exit(1)
	}

	r := remapper.NewMapper(opts.MemoryHog, opts.Verbose)

	err = r.RemapFrames(opts.InputDir, opts.InputPattern, opts.OutputDir, opts.StartIndex)
	if err != nil {
		panic(err)
	}
}
