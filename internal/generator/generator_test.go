package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateFromSchemaFile(t *testing.T) {
	outputDir := t.TempDir()

	err := Generate(Config{
		SchemaPath:  "../../test_data/example_1/schema.zed",
		OutputPath:  outputDir,
		PackageName: "permitions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that files were created for all definitions
	expectedFiles := []string{
		"bookingsvc_booking.go",
		"bookingsvc_brand.go",
		"bookingsvc_customer.go",
		"bookingsvc_employee.go",
		"bookingsvc_user.go",
		"menusvc_booking.go",
		"menusvc_user.go",
		"menusvc_company.go",
		"menusvc_order.go",
		"menusvc_table.go",
		"menusvc_customer.go",
		"menusvc_pricelist.go",
		"menusvc_product.go",
		"menusvc_setting.go",
		"client.go",
	}

	for _, name := range expectedFiles {
		path := filepath.Join(outputDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected file %s to be non-empty", name)
		}
	}
}

func TestGenerateFromSchemaFile_Example2(t *testing.T) {
	outputDir := t.TempDir()

	err := Generate(Config{
		SchemaPath:  "../../test_data/example_2/schema.zed",
		OutputPath:  outputDir,
		PackageName: "permissions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		"user.go",
		"anonymoususer.go",
		"platform.go",
		"forum.go",
		"post.go",
		"client.go",
	}

	for _, name := range expectedFiles {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestGenerateDefaultPackageName(t *testing.T) {
	// Use a named subdirectory to control the package name
	outputDir := filepath.Join(t.TempDir(), "mypkg")

	err := Generate(Config{
		SchemaPath: "../../test_data/example_2/schema.zed",
		OutputPath: outputDir,
		// PackageName deliberately empty — should default to "mypkg"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "user.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "package mypkg") {
		t.Errorf("expected 'package mypkg' in generated file, got:\n%s", string(content)[:200])
	}
}

func TestGenerateMissingSchemaFile(t *testing.T) {
	err := Generate(Config{
		SchemaPath:  "/nonexistent/path/schema.zed",
		OutputPath:  t.TempDir(),
		PackageName: "test",
	})
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
	if !strings.Contains(err.Error(), "reading schema") {
		t.Errorf("expected 'reading schema' in error, got: %v", err)
	}
}

func TestGenerateFromStringLexError(t *testing.T) {
	err := GenerateFromString("definition user { @ }", Config{
		OutputPath:  t.TempDir(),
		PackageName: "test",
	})
	if err == nil {
		t.Fatal("expected error for illegal token")
	}
	if !strings.Contains(err.Error(), "lexing schema") {
		t.Errorf("expected 'lexing schema' in error, got: %v", err)
	}
}

func TestGenerateFromStringParseError(t *testing.T) {
	// Valid tokens but invalid grammar
	err := GenerateFromString("definition {", Config{
		OutputPath:  t.TempDir(),
		PackageName: "test",
	})
	if err == nil {
		t.Fatal("expected error for invalid grammar")
	}
	if !strings.Contains(err.Error(), "parsing schema") {
		t.Errorf("expected 'parsing schema' in error, got: %v", err)
	}
}

func TestGenerateFromStringMkdirError(t *testing.T) {
	err := GenerateFromString("definition user {}", Config{
		OutputPath:  "/dev/null/impossible",
		PackageName: "test",
	})
	if err == nil {
		t.Fatal("expected error for invalid output path")
	}
	if !strings.Contains(err.Error(), "creating output directory") {
		t.Errorf("expected 'creating output directory' in error, got: %v", err)
	}
}

func TestGenerateFromStringWriteFileError(t *testing.T) {
	// Create a read-only directory
	dir := filepath.Join(t.TempDir(), "readonly")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	err := GenerateFromString("definition user {}", Config{
		OutputPath:  dir,
		PackageName: "test",
	})
	if err == nil {
		t.Fatal("expected error for read-only directory")
	}
	if !strings.Contains(err.Error(), "writing") {
		t.Errorf("expected 'writing' in error, got: %v", err)
	}
}

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mypkg", "mypkg"},
		{"MyPkg", "mypkg"},
		{"123pkg", "pkg_123pkg"},
		{"my-pkg", "my_pkg"},
		{"my.pkg", "my_pkg"},
		{"", "generated"},
		{"my pkg", "my_pkg"},
		{"_private", "_private"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizePackageName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePackageName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
