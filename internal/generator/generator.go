package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/oitnes/authzed-codegen/internal/generator/codegen"
	"github.com/oitnes/authzed-codegen/internal/generator/parser"
	zedlexer "github.com/oitnes/authzed-codegen/internal/generator/zed_lexer"
)

// Config holds the configuration for the code generation pipeline.
type Config struct {
	SchemaPath     string
	OutputPath     string
	PackageName    string
	WithRepository bool
	CleanPackage   bool
}

// Generate runs the full pipeline: read schema → lex → parse → generate → write.
func Generate(cfg Config) error {
	content, err := os.ReadFile(cfg.SchemaPath)
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	return GenerateFromString(string(content), cfg)
}

// GenerateFromString runs the pipeline from a schema string.
func GenerateFromString(schemaContent string, cfg Config) error {
	tokens, err := zedlexer.Lex(schemaContent)
	if err != nil {
		return fmt.Errorf("lexing schema: %w", err)
	}

	schema, err := parser.Parse(tokens)
	if err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	packageName := cfg.PackageName
	if packageName == "" {
		packageName = sanitizePackageName(filepath.Base(cfg.OutputPath))
	}

	files, err := codegen.Generate(schema, codegen.Options{
		PackageName:    packageName,
		WithRepository: cfg.WithRepository,
	})
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	if cfg.CleanPackage {
		if err := os.RemoveAll(cfg.OutputPath); err != nil {
			return fmt.Errorf("removing output directory: %w", err)
		}
	}

	if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, file := range files {
		filePath := filepath.Join(cfg.OutputPath, file.Name)
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", file.Name, err)
		}
	}

	return nil
}

// sanitizePackageName ensures the directory name is a valid Go package name.
func sanitizePackageName(name string) string {
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}
		return '_'
	}, name)
	// Package name must start with a letter or underscore
	if len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = "pkg_" + name
	}
	if name == "" {
		name = "generated"
	}
	return name
}
