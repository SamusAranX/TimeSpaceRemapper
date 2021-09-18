package main

type CommandLineOpts struct {
	InputDir     string `short:"i" long:"input-dir" description:"Input frame directory"`
	OutputDir    string `short:"o" long:"output-dir" description:"Output frame directory"`
	InputPattern string `short:"p" long:"pattern" description:"Input file name pattern (takes glob-style patterns)"`

	StartIndex int `short:"s" long:"start-index" default:"0" description:"Starting index"`

	MemoryHog bool `short:"M" long:"memory-hog" description:"Hog Mode (will attempt to keep all new frames in memory)"`

	NoHeader bool `long:"no-header" description:"Hide header"`
	Version  bool `short:"v" long:"version" description:"Show version and exit"`

	//Positional struct {
	//	OutputFile string `positional-arg-name:"OUTPUT FILE"`
	//} `positional-args:"true"`
}
