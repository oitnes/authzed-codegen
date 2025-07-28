package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oitnes/authzed-codegen/internal/generator"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var (
		outputPath  string
		schemePath  string
		showVersion bool
		showHelp    bool
	)

	flag.StringVar(&outputPath, "output", "", "output path for generated files")
	flag.StringVar(&schemePath, "schema", "", "input zed schema file path")
	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.BoolVar(&showVersion, "v", false, "show version information (short)")
	flag.BoolVar(&showHelp, "help", false, "show help message")
	flag.BoolVar(&showHelp, "h", false, "show help message (short)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "authzed-codegen - Type-safe Go code generator for SpiceDB schemas\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --schema schema.zed --output ./generated\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -schema examples/example_1.zed -output ./output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFor more information, visit: https://github.com/oitnes/authzed-codegen\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("authzed-codegen %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if schemePath == "" {
		fmt.Print("missing schema path")
		os.Exit(1)
	}

	if outputPath == "" {
		fmt.Print("missing output path")
		os.Exit(1)
	}

	input, err := os.ReadFile(schemePath)
	if err != nil {
		fmt.Printf("error during reading schema file: %e", err)
		os.Exit(1)
	}

	err = generator.GenerateCodeFromString(string(input), outputPath)
	if err != nil {
		fmt.Printf("error code generation: %e", err)
		os.Exit(1)
	}
}
