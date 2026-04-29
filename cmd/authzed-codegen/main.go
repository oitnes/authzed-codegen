package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oitnes/authzed-codegen/internal/generator"
)

func main() {
	var cfg generator.Config

	flag.StringVar(&cfg.SchemaPath, "schema", "", "path to .zed schema file (required)")
	flag.StringVar(&cfg.OutputPath, "output", "", "output directory for generated Go files (required)")
	flag.StringVar(&cfg.PackageName, "package", "", "package name for generated code (defaults to output directory name)")
	flag.BoolVar(&cfg.WithRepository, "with-repository", false, "generate entity CRUD methods")
	flag.BoolVar(&cfg.CleanPackage, "clean-package", false, "remove output directory before generating code")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "authzed-codegen - Type-safe Go code generator for SpiceDB schemas\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --schema schema.zed --output ./permissions\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --schema schema.zed --output ./permissions --with-repository\n", os.Args[0])
	}

	flag.Parse()

	if cfg.SchemaPath == "" || cfg.OutputPath == "" {
		fmt.Fprintln(os.Stderr, "error: --schema and --output are required")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.PackageName == "" {
		cfg.PackageName = filepath.Base(cfg.OutputPath)
	}

	if err := generator.Generate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
