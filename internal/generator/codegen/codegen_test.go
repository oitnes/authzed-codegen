package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/oitnes/authzed-codegen/internal/generator/ast"
)

func TestGenerateEmptyDefinition(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	var userFile *GeneratedFile
	for _, f := range files {
		if f.Name == "user.go" {
			userFile = f
		}
	}
	if userFile == nil {
		t.Fatal("expected user.go file")
	}

	assertValidGo(t, userFile)
	assertContains(t, userFile.Content, "TypeUser")
	assertContains(t, userFile.Content, "type User struct")
	assertContains(t, userFile.Content, "func NewUser")
}

func TestGenerateDefinitionWithRelations(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "document",
				Relations: []*ast.Relation{
					{
						Name: "owner",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user"},
						},
					},
				},
			},
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	// Find document file
	var docFile *GeneratedFile
	for _, f := range files {
		if f.Name == "document.go" {
			docFile = f
		}
	}
	if docFile == nil {
		t.Fatal("expected document.go file")
	}

	assertValidGo(t, docFile)
	assertContains(t, docFile.Content, "DocumentRelationOwner")
	assertContains(t, docFile.Content, "DocumentOwnerObjects")
	assertContains(t, docFile.Content, "CreateOwnerRelations")
	assertContains(t, docFile.Content, "ReadOwnerRelations")
	assertContains(t, docFile.Content, "DeleteOwnerRelations")
}

func TestGenerateDefinitionWithPermissions(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "document",
				Relations: []*ast.Relation{
					{
						Name: "owner",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user"},
						},
					},
				},
				Permissions: []*ast.Permission{
					{
						Name:       "edit",
						Expression: &ast.RelationRef{Name: "owner"},
					},
				},
			},
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var docFile *GeneratedFile
	for _, f := range files {
		if f.Name == "document.go" {
			docFile = f
		}
	}
	if docFile == nil {
		t.Fatal("expected document.go file")
	}

	assertValidGo(t, docFile)
	assertContains(t, docFile.Content, "DocumentPermissionEdit")
	assertContains(t, docFile.Content, "CheckDocumentEditInputs")
	assertContains(t, docFile.Content, "CheckEdit")
	assertContains(t, docFile.Content, "LookupDocumentsWith")
}

func TestGenerateNamespacedDefinition(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "bookingsvc/booking",
				Relations: []*ast.Relation{
					{
						Name: "owner",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "bookingsvc/employee"},
						},
					},
				},
			},
			{Name: "bookingsvc/employee"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "permitions"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bookingFile *GeneratedFile
	for _, f := range files {
		if f.Name == "bookingsvc_booking.go" {
			bookingFile = f
		}
	}
	if bookingFile == nil {
		t.Fatal("expected bookingsvc_booking.go file")
	}

	assertValidGo(t, bookingFile)
	assertContains(t, bookingFile.Content, "TypeBookingsvcBooking")
	assertContains(t, bookingFile.Content, "BookingsvcBooking")
	assertContains(t, bookingFile.Content, "NewBookingsvcBooking")
}

func TestGenerateWithRepository(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz", WithRepository: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var userFile *GeneratedFile
	for _, f := range files {
		if f.Name == "user.go" {
			userFile = f
		}
	}
	if userFile == nil {
		t.Fatal("expected user.go file")
	}

	assertValidGo(t, userFile)
	assertContains(t, userFile.Content, "repo")
	assertContains(t, userFile.Content, "CreateUser")
	assertContains(t, userFile.Content, "GetUser")
	assertContains(t, userFile.Content, "Update")
	assertContains(t, userFile.Content, "Delete")
	assertContains(t, userFile.Content, "CheckUserExists")
	assertContains(t, userFile.Content, "ListUsers")
}

func TestGenerateWithRepositoryRelationsUseRepoConstructor(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "document",
				Relations: []*ast.Relation{
					{
						Name: "owner",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user"},
						},
					},
				},
			},
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz", WithRepository: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var docFile *GeneratedFile
	for _, f := range files {
		if f.Name == "document.go" {
			docFile = f
		}
	}
	if docFile == nil {
		t.Fatal("expected document.go file")
	}

	assertValidGo(t, docFile)
	assertContains(t, docFile.Content, "NewUser(string(id), d.engine, d.repo)")
}

func TestGenerateWithRepositoryPermissionsUseRepoConstructor(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "document",
				Relations: []*ast.Relation{
					{
						Name: "owner",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user"},
						},
					},
				},
				Permissions: []*ast.Permission{
					{
						Name:       "view",
						Expression: &ast.RelationRef{Name: "owner"},
					},
				},
			},
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz", WithRepository: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var docFile *GeneratedFile
	for _, f := range files {
		if f.Name == "document.go" {
			docFile = f
		}
	}
	if docFile == nil {
		t.Fatal("expected document.go file")
	}

	assertValidGo(t, docFile)
	assertContains(t, docFile.Content, "LookupDocumentsWithViewByUser(ctx context.Context, engine authz.Engine, subject User, repo authz.Repository)")
	assertContains(t, docFile.Content, "NewDocument(string(id), engine, repo)")
	assertContains(t, docFile.Content, "NewUser(string(id), d.engine, d.repo)")
}

func TestGenerateWildcardRelationSupport(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "platform",
				Relations: []*ast.Relation{
					{
						Name: "anonim_user",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user", IsWildcard: true},
						},
					},
				},
			},
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var platformFile *GeneratedFile
	for _, f := range files {
		if f.Name == "platform.go" {
			platformFile = f
		}
	}
	if platformFile == nil {
		t.Fatal("expected platform.go file")
	}

	assertValidGo(t, platformFile)
	assertContains(t, platformFile.Content, "UserWildcard bool")
	assertContains(t, platformFile.Content, "if subjects.UserWildcard")
	assertContains(t, platformFile.Content, "authz.ID(\"*\")")
	assertContains(t, platformFile.Content, "if id == authz.ID(\"*\")")
}

func TestGenerateWithoutRepository(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz", WithRepository: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertValidGo(t, files[0])
	assertNotContains(t, files[0].Content, "repo")
	assertNotContains(t, files[0].Content, "CreateUser(")
	assertNotContains(t, files[0].Content, "Repository")
}

func TestGenerateMultipleSubjectTypes(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{
				Name: "document",
				Relations: []*ast.Relation{
					{
						Name: "editor",
						SubjectTypes: []*ast.SubjectType{
							{TypeName: "user"},
							{TypeName: "group"},
						},
					},
				},
			},
			{Name: "user"},
			{Name: "group"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var docFile *GeneratedFile
	for _, f := range files {
		if f.Name == "document.go" {
			docFile = f
		}
	}
	if docFile == nil {
		t.Fatal("expected document.go file")
	}

	assertValidGo(t, docFile)
	assertContains(t, docFile.Content, "User")
	assertContains(t, docFile.Content, "[]User")
	assertContains(t, docFile.Content, "Group")
	assertContains(t, docFile.Content, "[]Group")
}

func TestGenerateDoNotEditHeader(t *testing.T) {
	schema := &ast.Schema{
		Definitions: []*ast.Definition{
			{Name: "user"},
		},
	}

	files, err := Generate(schema, Options{PackageName: "authz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, files[0].Content, "Code generated by authzed-codegen. DO NOT EDIT.")
}

// assertValidGo verifies the generated content is syntactically valid Go.
func assertValidGo(t *testing.T, file *GeneratedFile) {
	t.Helper()
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, file.Name, file.Content, parser.AllErrors)
	if err != nil {
		t.Errorf("generated file %s is not valid Go:\n%v\n\nContent:\n%s", file.Name, err, file.Content)
	}
}

func assertContains(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("expected content to contain %q, but it doesn't.\nContent:\n%s", substr, content)
	}
}

func assertNotContains(t *testing.T, content, substr string) {
	t.Helper()
	if strings.Contains(content, substr) {
		t.Errorf("expected content NOT to contain %q, but it does", substr)
	}
}
