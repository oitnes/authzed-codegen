package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oitnes/authzed-codegen/internal/generator/zed_lexer"
)

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"h", "H"},
		{"hello_world", "Hello_world"},
	}

	for _, test := range tests {
		result := UpperFirst(test.input)
		if result != test.expected {
			t.Errorf("UpperFirst(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello-world", "helloworld"},
		{"hello_world", "helloworld"},
		{"Hello-World_Test", "helloworldtest"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, test := range tests {
		result := PackageName(test.input)
		if result != test.expected {
			t.Errorf("PackageName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSnakeToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"test_case", "TestCase"},
		{"single", "Single"},
		{"", ""},
		{"snake_case_example", "SnakeCaseExample"},
	}

	for _, test := range tests {
		result := SnakeToPascal(test.input)
		if result != test.expected {
			t.Errorf("SnakeToPascal(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"platform/user", "User"},
		{"namespace/platform/user", "Platform"},
		{"", ""},
		{"hello-world", "Helloworld"},
	}

	for _, test := range tests {
		result := TypeName(test.input)
		if result != test.expected {
			t.Errorf("TypeName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestTypeNameWithUnderscores(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"private_forum", "Private_forum"},
		{"platform/private_forum", "Private_forum"},
		{"namespace/platform/public_forum", "Public_forum"},
		{"", ""},
	}

	for _, test := range tests {
		result := TypeNameWithUnderscores(test.input)
		if result != test.expected {
			t.Errorf("TypeNameWithUnderscores(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestNewParser(t *testing.T) {
	input := "definition user {}"
	parser := NewParser(input)
	if parser == nil {
		t.Error("NewParser should return a non-nil parser")
	}
}

func TestFilterComments(t *testing.T) {
	tokens := []zed_lexer.Token{
		{Type: zed_lexer.DEFINITION, Literal: "definition"},
		{Type: zed_lexer.COMMENT, Literal: "// comment"},
		{Type: zed_lexer.IDENTIFIER, Literal: "user"},
		{Type: zed_lexer.COMMENT, Literal: "/* block comment */"},
		{Type: zed_lexer.LBRACE, Literal: "{"},
	}

	filtered := filterComments(tokens)
	expectedCount := 3 // definition, identifier, lbrace

	if len(filtered) != expectedCount {
		t.Errorf("filterComments should return %d tokens, got %d", expectedCount, len(filtered))
	}

	for _, token := range filtered {
		if token.Type == zed_lexer.COMMENT {
			t.Error("filterComments should remove all comment tokens")
		}
	}
}

func TestParseSimpleDefinition(t *testing.T) {
	input := "definition user {}"
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(definitions))
	}

	def := definitions[0]
	if def.Name != "user" {
		t.Errorf("Expected definition name 'user', got '%s'", def.Name)
	}

	if len(def.Relations) != 0 {
		t.Errorf("Expected 0 relations, got %d", len(def.Relations))
	}

	if len(def.Permissions) != 0 {
		t.Errorf("Expected 0 permissions, got %d", len(def.Permissions))
	}
}

func TestParseDefinitionWithRelation(t *testing.T) {
	input := `definition forum {
		relation owner: user
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(definitions))
	}

	def := definitions[0]
	if def.Name != "forum" {
		t.Errorf("Expected definition name 'forum', got '%s'", def.Name)
	}

	if len(def.Relations) != 1 {
		t.Fatalf("Expected 1 relation, got %d", len(def.Relations))
	}

	rel := def.Relations[0]
	if rel.Name != "owner" {
		t.Errorf("Expected relation name 'owner', got '%s'", rel.Name)
	}

	if rel.Expression.String() != "user" {
		t.Errorf("Expected relation expression 'user', got '%s'", rel.Expression.String())
	}
}

func TestParseDefinitionWithPermission(t *testing.T) {
	input := `definition forum {
		relation owner: user
		permission edit = owner
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	if len(def.Permissions) != 1 {
		t.Fatalf("Expected 1 permission, got %d", len(def.Permissions))
	}

	perm := def.Permissions[0]
	if perm.Name != "edit" {
		t.Errorf("Expected permission name 'edit', got '%s'", perm.Name)
	}

	if perm.Expression.String() != "owner" {
		t.Errorf("Expected permission expression 'owner', got '%s'", perm.Expression.String())
	}
}

func TestParseComplexPermissionExpression(t *testing.T) {
	input := `definition forum {
		relation owner: user
		relation member: user
		relation banned: user
		permission view = owner + member - banned
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	perm := def.Permissions[0]
	expected := "owner + member - banned"
	if perm.Expression.String() != expected {
		t.Errorf("Expected permission expression '%s', got '%s'", expected, perm.Expression.String())
	}
}

func TestParsePermissionWithArrow(t *testing.T) {
	input := `definition forum {
		relation global: platform
		permission super_admin = global->admin
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	perm := def.Permissions[0]
	expected := "global->admin"
	if perm.Expression.String() != expected {
		t.Errorf("Expected permission expression '%s', got '%s'", expected, perm.Expression.String())
	}
}

func TestParsePermissionWithParentheses(t *testing.T) {
	input := `definition forum {
		relation owner: user
		relation admin: user
		relation member: user
		relation banned: user
		permission make_post = (owner + admin + member) - banned
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	perm := def.Permissions[0]
	// The exact string representation may vary based on parsing, but should contain the essential elements
	exprStr := perm.Expression.String()
	if !strings.Contains(exprStr, "owner") || !strings.Contains(exprStr, "admin") || !strings.Contains(exprStr, "member") || !strings.Contains(exprStr, "banned") {
		t.Errorf("Expected permission expression to contain all identifiers, got '%s'", exprStr)
	}
}

func TestParseRelationWithUnion(t *testing.T) {
	input := `definition post {
		relation location: private_forum | public_forum
	}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	rel := def.Relations[0]
	expected := "private_forum | public_forum"
	if rel.Expression.String() != expected {
		t.Errorf("Expected relation expression '%s', got '%s'", expected, rel.Expression.String())
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing definition keyword", "user {}"},
		{"missing braces", "definition user"},
		{"missing colon in relation", "definition user { relation owner user }"},
		{"missing equal in permission", "definition user { permission edit owner }"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(test.input)
			_, err := parser.ParseDefinitions()
			if err == nil {
				t.Errorf("Expected parsing error for input: %s", test.input)
			}
		})
	}
}

func TestFlattenRelationExpressionTypes(t *testing.T) {
	// Test SingleRelation
	singleRel := &SingleRelation{Value: "user"}
	relations := flattenRelationExpressionTypes(singleRel)
	if len(relations) != 1 || relations[0].Value != "user" {
		t.Errorf("Expected single relation 'user', got %v", relations)
	}

	// Test UnionRelation
	unionRel := &UnionRelation{
		Left:  &SingleRelation{Value: "private_forum"},
		Right: &SingleRelation{Value: "public_forum"},
	}
	relations = flattenRelationExpressionTypes(unionRel)
	if len(relations) != 2 {
		t.Errorf("Expected 2 relations from union, got %d", len(relations))
	}
}

func TestFlattenRelationExpressionTypeStrings(t *testing.T) {
	singleRel := &SingleRelation{Value: "user"}
	types := flattenRelationExpressionTypeStrings(singleRel)
	if len(types) != 1 || types[0] != "user" {
		t.Errorf("Expected ['user'], got %v", types)
	}
}

func TestGenerateCodeFromString(t *testing.T) {
	input := `definition user {}`

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "generator_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = GenerateCodeFromString(input, tempDir)
	if err != nil {
		t.Fatalf("GenerateCodeFromString failed: %v", err)
	}

	// Check if file was created
	expectedFile := filepath.Join(tempDir, "user_gen.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", expectedFile)
	}

	// Read and verify file content
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Error("Generated file should contain 'package main'")
	}
	if !strings.Contains(contentStr, "const TypeUser") {
		t.Error("Generated file should contain 'const TypeUser'")
	}
}

func TestGenerateCodeFromComplexString(t *testing.T) {
	input := `definition platform {
		relation admin: user
		permission super_admin = admin
	}
	
	definition forum {
		relation global: platform
		relation owner: user
		permission edit = owner + global->super_admin
	}`

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "generator_test_complex")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = GenerateCodeFromString(input, tempDir)
	if err != nil {
		t.Fatalf("GenerateCodeFromString failed: %v", err)
	}

	// Check if both files were created
	platformFile := filepath.Join(tempDir, "platform_gen.go")
	forumFile := filepath.Join(tempDir, "forum_gen.go")

	if _, err := os.Stat(platformFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", platformFile)
	}
	if _, err := os.Stat(forumFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", forumFile)
	}
}

func TestParseDefinitionWithPrefixes(t *testing.T) {
	input := `definition namespace/platform/user {}`
	parser := NewParser(input)
	definitions, err := parser.ParseDefinitions()

	if err != nil {
		t.Fatalf("ParseDefinitions failed: %v", err)
	}

	def := definitions[0]
	if def.Name != "user" {
		t.Errorf("Expected definition name 'user', got '%s'", def.Name)
	}

	expectedPrefixes := []string{"namespace", "platform"}
	if len(def.Prefixes) != len(expectedPrefixes) {
		t.Errorf("Expected %d prefixes, got %d", len(expectedPrefixes), len(def.Prefixes))
	}

	for i, prefix := range expectedPrefixes {
		if def.Prefixes[i] != prefix {
			t.Errorf("Expected prefix[%d] = '%s', got '%s'", i, prefix, def.Prefixes[i])
		}
	}
}

func TestObjectTypeNodeString(t *testing.T) {
	tests := []struct {
		name     string
		obj      ObjectTypeNode
		expected string
	}{
		{"no prefix", ObjectTypeNode{Name: "user", Prefix: ""}, "user"},
		{"with prefix", ObjectTypeNode{Name: "user", Prefix: "platform"}, "platform/user"},
		{"with nested prefix", ObjectTypeNode{Name: "user", Prefix: "namespace/platform"}, "namespace/platform/user"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.obj.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestDefinitionGetObjectType(t *testing.T) {
	def := Definition{
		Name:     "user",
		Prefixes: []string{"namespace", "platform"},
	}

	objType := def.GetObjectType()
	if objType.Name != "user" {
		t.Errorf("Expected name 'user', got '%s'", objType.Name)
	}
	if objType.Prefix != "namespace/platform" {
		t.Errorf("Expected prefix 'namespace/platform', got '%s'", objType.Prefix)
	}
}

func TestBinaryOpNodeString(t *testing.T) {
	tests := []struct {
		name     string
		node     BinaryOpNode
		expected string
	}{
		{
			"arrow operator",
			BinaryOpNode{
				Operator: "->",
				Left:     &IdentifierNode{Value: "global"},
				Right:    &IdentifierNode{Value: "admin"},
			},
			"global->admin",
		},
		{
			"plus operator",
			BinaryOpNode{
				Operator: "+",
				Left:     &IdentifierNode{Value: "owner"},
				Right:    &IdentifierNode{Value: "admin"},
			},
			"owner + admin",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.node.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestTransformDefinitions(t *testing.T) {
	definitions := []Definition{
		{
			Name: "user",
			Relations: []RelationNode{
				{Name: "friend", Expression: &SingleRelation{Value: "user"}},
			},
			Permissions: []PermissionNode{
				{Name: "edit", Expression: &IdentifierNode{Value: "friend"}},
			},
		},
	}

	transformed := transformDefinitions(definitions)
	if len(transformed) != 1 {
		t.Errorf("Expected 1 transformed definition, got %d", len(transformed))
	}

	userDef := transformed["user"]
	if userDef == nil {
		t.Fatal("Expected user definition to exist")
	}

	if len(userDef.Relations) != 1 {
		t.Errorf("Expected 1 relation, got %d", len(userDef.Relations))
	}

	if len(userDef.Permissions) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(userDef.Permissions))
	}
}
