package parser

import (
	"testing"

	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	zedlexer "github.com/oitnes/authzed-codegen/internal/generator/zed_lexer"
)

func mustLex(t *testing.T, input string) []zedlexer.Token {
	t.Helper()
	tokens, err := zedlexer.Lex(input)
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	return tokens
}

func TestParseEmptySchema(t *testing.T) {
	tokens := mustLex(t, "")
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(schema.Definitions))
	}
}

func TestParseEmptyDefinition(t *testing.T) {
	tokens := mustLex(t, `definition user {}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(schema.Definitions))
	}
	if schema.Definitions[0].Name != "user" {
		t.Errorf("expected name 'user', got %q", schema.Definitions[0].Name)
	}
	if len(schema.Definitions[0].Relations) != 0 {
		t.Errorf("expected 0 relations, got %d", len(schema.Definitions[0].Relations))
	}
}

func TestParseNamespacedDefinition(t *testing.T) {
	tokens := mustLex(t, `definition bookingsvc/booking {}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Definitions[0].Name != "bookingsvc/booking" {
		t.Errorf("expected name 'bookingsvc/booking', got %q", schema.Definitions[0].Name)
	}
}

func TestParseRelation(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation owner: user
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := schema.Definitions[0]
	if len(def.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(def.Relations))
	}
	rel := def.Relations[0]
	if rel.Name != "owner" {
		t.Errorf("expected relation name 'owner', got %q", rel.Name)
	}
	if len(rel.SubjectTypes) != 1 {
		t.Fatalf("expected 1 subject type, got %d", len(rel.SubjectTypes))
	}
	if rel.SubjectTypes[0].TypeName != "user" {
		t.Errorf("expected subject type 'user', got %q", rel.SubjectTypes[0].TypeName)
	}
	if rel.SubjectTypes[0].IsWildcard {
		t.Error("expected IsWildcard=false")
	}
}

func TestParseRelationWithUnionTypes(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation creator: employee | customer
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rel := schema.Definitions[0].Relations[0]
	if len(rel.SubjectTypes) != 2 {
		t.Fatalf("expected 2 subject types, got %d", len(rel.SubjectTypes))
	}
	if rel.SubjectTypes[0].TypeName != "employee" {
		t.Errorf("expected first subject 'employee', got %q", rel.SubjectTypes[0].TypeName)
	}
	if rel.SubjectTypes[1].TypeName != "customer" {
		t.Errorf("expected second subject 'customer', got %q", rel.SubjectTypes[1].TypeName)
	}
}

func TestParseRelationWithWildcard(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation viewer: user:*
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rel := schema.Definitions[0].Relations[0]
	if !rel.SubjectTypes[0].IsWildcard {
		t.Error("expected IsWildcard=true")
	}
}

func TestParseSimplePermission(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation owner: user
		permission edit = owner
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	perm := schema.Definitions[0].Permissions[0]
	if perm.Name != "edit" {
		t.Errorf("expected permission name 'edit', got %q", perm.Name)
	}
	ref, ok := perm.Expression.(*ast.RelationRef)
	if !ok {
		t.Fatalf("expected RelationRef, got %T", perm.Expression)
	}
	if ref.Name != "owner" {
		t.Errorf("expected relation ref 'owner', got %q", ref.Name)
	}
}

func TestParseUnionExpression(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation owner: user
		relation editor: user
		permission edit = owner + editor
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expr, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.UnionExpr)
	if !ok {
		t.Fatalf("expected UnionExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	assertRelRef(t, expr.Left, "owner")
	assertRelRef(t, expr.Right, "editor")
}

func TestParseArrowExpression(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation parent: folder
		permission view = parent->view
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arrow, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.ArrowExpr)
	if !ok {
		t.Fatalf("expected ArrowExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	if arrow.Relation != "parent" {
		t.Errorf("expected arrow relation 'parent', got %q", arrow.Relation)
	}
	if arrow.Permission != "view" {
		t.Errorf("expected arrow permission 'view', got %q", arrow.Permission)
	}
}

func TestParseExclusionExpression(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation member: user
		relation banned: user
		permission view = member - banned
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	excl, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.ExclusionExpr)
	if !ok {
		t.Fatalf("expected ExclusionExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	assertRelRef(t, excl.Left, "member")
	assertRelRef(t, excl.Right, "banned")
}

func TestParseIntersectionExpression(t *testing.T) {
	tokens := mustLex(t, `definition doc {
		relation member: user
		relation signed_tos: user
		permission access = member & signed_tos
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inter, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.IntersectionExpr)
	if !ok {
		t.Fatalf("expected IntersectionExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	assertRelRef(t, inter.Left, "member")
	assertRelRef(t, inter.Right, "signed_tos")
}

func TestParseGroupedExpression(t *testing.T) {
	// (admin + member) - banned
	tokens := mustLex(t, `definition forum {
		relation admin: user
		relation member: user
		relation banned: user
		permission post = (admin + member) - banned
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	excl, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.ExclusionExpr)
	if !ok {
		t.Fatalf("expected ExclusionExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	union, ok := excl.Left.(*ast.UnionExpr)
	if !ok {
		t.Fatalf("expected UnionExpr in left of exclusion, got %T", excl.Left)
	}
	assertRelRef(t, union.Left, "admin")
	assertRelRef(t, union.Right, "member")
	assertRelRef(t, excl.Right, "banned")
}

func TestParseComplexNestedExpression(t *testing.T) {
	// owner + (((admin + member) & signed_tos) - banned)
	tokens := mustLex(t, `definition private_forum {
		relation owner: user
		relation admin: user
		relation member: user
		relation banned: user
		relation signed_tos: user
		permission make_post = owner + (((admin + member) & signed_tos) - banned)
	}`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Top level: union of owner + (...)
	topUnion, ok := schema.Definitions[0].Permissions[0].Expression.(*ast.UnionExpr)
	if !ok {
		t.Fatalf("expected top-level UnionExpr, got %T", schema.Definitions[0].Permissions[0].Expression)
	}
	assertRelRef(t, topUnion.Left, "owner")

	// Right side: (...) - banned
	excl, ok := topUnion.Right.(*ast.ExclusionExpr)
	if !ok {
		t.Fatalf("expected ExclusionExpr, got %T", topUnion.Right)
	}
	assertRelRef(t, excl.Right, "banned")

	// Left of exclusion: (...) & signed_tos
	inter, ok := excl.Left.(*ast.IntersectionExpr)
	if !ok {
		t.Fatalf("expected IntersectionExpr, got %T", excl.Left)
	}
	assertRelRef(t, inter.Right, "signed_tos")

	// Left of intersection: admin + member
	union, ok := inter.Left.(*ast.UnionExpr)
	if !ok {
		t.Fatalf("expected UnionExpr, got %T", inter.Left)
	}
	assertRelRef(t, union.Left, "admin")
	assertRelRef(t, union.Right, "member")
}

func TestParseMultipleDefinitions(t *testing.T) {
	tokens := mustLex(t, `
		definition user {}
		definition document {
			relation owner: user
			permission edit = owner
		}
	`)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(schema.Definitions))
	}
	if schema.Definitions[0].Name != "user" {
		t.Errorf("expected first def name 'user', got %q", schema.Definitions[0].Name)
	}
	if schema.Definitions[1].Name != "document" {
		t.Errorf("expected second def name 'document', got %q", schema.Definitions[1].Name)
	}
}

func TestParseExample1Schema(t *testing.T) {
	input := `
definition bookingsvc/booking {
	relation owner: bookingsvc/employee
	relation creator: bookingsvc/employee | bookingsvc/customer
	permission write = creator + owner + owner->manage + creator->manage
	permission change_owner = creator + creator->manage
}

definition bookingsvc/brand {
	relation admin: bookingsvc/user
	relation manager: bookingsvc/employee
	relation employee: bookingsvc/employee
	permission manage = manager + admin
	permission create_booking = manage + employee
}

definition bookingsvc/customer {}

definition bookingsvc/employee {
	relation account: bookingsvc/user
	relation belongs_brand: bookingsvc/brand
	relation viewer: bookingsvc/user:*
	permission manage = account + belongs_brand->manage
	permission view = manage + viewer
}

definition bookingsvc/user {}
`

	tokens := mustLex(t, input)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Definitions) != 5 {
		t.Fatalf("expected 5 definitions, got %d", len(schema.Definitions))
	}

	// Check bookingsvc/booking
	booking := schema.Definitions[0]
	if booking.Name != "bookingsvc/booking" {
		t.Errorf("expected 'bookingsvc/booking', got %q", booking.Name)
	}
	if len(booking.Relations) != 2 {
		t.Errorf("expected 2 relations, got %d", len(booking.Relations))
	}
	if len(booking.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(booking.Permissions))
	}

	// Check creator relation has 2 subject types
	creator := booking.Relations[1]
	if creator.Name != "creator" {
		t.Errorf("expected relation 'creator', got %q", creator.Name)
	}
	if len(creator.SubjectTypes) != 2 {
		t.Errorf("expected 2 subject types for creator, got %d", len(creator.SubjectTypes))
	}

	// Check wildcard relation in employee
	employee := schema.Definitions[3]
	viewer := employee.Relations[2]
	if !viewer.SubjectTypes[0].IsWildcard {
		t.Error("expected viewer to be a wildcard relation")
	}

	// Check empty definitions
	customer := schema.Definitions[2]
	if len(customer.Relations) != 0 || len(customer.Permissions) != 0 {
		t.Error("expected empty definition for customer")
	}
}

func TestParseExample2Schema(t *testing.T) {
	input := `
definition user {}

definition platform {
	relation administrator: user
	relation registered_user: user
	relation anonim_user: user:*
	permission super_admin = administrator
	permission create_forum = registered_user
	permission subscribe_forums = registered_user
	permission view = anonim_user + registered_user + administrator
}

definition public_forum {
	relation global: platform
	relation owner: user
	relation admin: user
	relation member: user
	relation banned: user
	permission own = owner
	permission edit_admin_list = own + global->super_admin
	permission delete = own + global->super_admin
	permission edit = owner + admin + global->super_admin
	permission view = edit + member + global->view - banned
	permission make_post = (owner + admin + member) - banned
}

definition private_forum {
	relation global: platform
	relation owner: user
	relation admin: user
	relation member: user
	relation aplicant: user
	relation invited_candidate: user
	relation banned: user
	relation signed_tos: user
	permission own = owner
	permission edit_admin_list = own + global->super_admin
	permission delete = own + global->super_admin
	permission apply = global->subscribe_forums - banned
	permission edit = owner + admin + global->super_admin
	permission make_post = owner + (((admin + member) & signed_tos) - banned)
	permission view = make_post + invited_candidate + aplicant + global->super_admin
}

definition post {
	relation location: private_forum | public_forum
	relation author: user
	permission edit = author + location->edit
	permission view = edit + location->view
	permission delete = edit
}
`

	tokens := mustLex(t, input)
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Definitions) != 5 {
		t.Fatalf("expected 5 definitions, got %d", len(schema.Definitions))
	}

	// Check private_forum make_post: owner + (((admin + member) & signed_tos) - banned)
	privateForum := schema.Definitions[3]
	if privateForum.Name != "private_forum" {
		t.Fatalf("expected 'private_forum', got %q", privateForum.Name)
	}

	makePost := privateForum.Permissions[5]
	if makePost.Name != "make_post" {
		t.Fatalf("expected permission 'make_post', got %q", makePost.Name)
	}

	// Top: union of owner + (...)
	topUnion, ok := makePost.Expression.(*ast.UnionExpr)
	if !ok {
		t.Fatalf("expected UnionExpr at top, got %T", makePost.Expression)
	}
	assertRelRef(t, topUnion.Left, "owner")

	// post definition has location with union types
	post := schema.Definitions[4]
	location := post.Relations[0]
	if len(location.SubjectTypes) != 2 {
		t.Errorf("expected 2 subject types for location, got %d", len(location.SubjectTypes))
	}
}

func TestParseErrorMissingBrace(t *testing.T) {
	tokens := mustLex(t, `definition user {`)
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if parseErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestParseErrorInvalidToken(t *testing.T) {
	tokens := mustLex(t, `definition user {
		relation owner =
	}`)
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertRelRef(t *testing.T, expr ast.Expr, name string) {
	t.Helper()
	ref, ok := expr.(*ast.RelationRef)
	if !ok {
		t.Fatalf("expected RelationRef, got %T", expr)
	}
	if ref.Name != name {
		t.Errorf("expected relation ref %q, got %q", name, ref.Name)
	}
}

func tok(tt zedlexer.TokenType, lit string) zedlexer.Token {
	return zedlexer.Token{Type: tt, Literal: lit, Line: 1, Column: 1}
}

func TestParseErrorError(t *testing.T) {
	e := &ParseError{Line: 5, Column: 10, Message: "something wrong"}
	got := e.Error()
	want := "parse error at line 5, column 10: something wrong"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestParseSchemaUnexpectedToken(t *testing.T) {
	// A token that is neither DEFINITION, CAVEAT, nor EOF at the schema level
	tokens := []zedlexer.Token{
		tok(zedlexer.IDENTIFIER, "oops"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for unexpected token at schema level")
	}
}

func TestParseSchemaEOFToken(t *testing.T) {
	// Schema with just an EOF token
	tokens := []zedlexer.Token{
		tok(zedlexer.EOF, ""),
	}
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(schema.Definitions))
	}
}

func TestParseSkipCaveat(t *testing.T) {
	// Construct tokens manually since the lexer can't handle caveat body syntax
	tokens := []zedlexer.Token{
		tok(zedlexer.CAVEAT, "caveat"),
		tok(zedlexer.IDENTIFIER, "some_caveat"),
		tok(zedlexer.IDENTIFIER, "param"),
		tok(zedlexer.IDENTIFIER, "int"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.IDENTIFIER, "param"),
		tok(zedlexer.RBRACE, "}"),
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "user"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RBRACE, "}"),
	}
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 1 {
		t.Fatalf("expected 1 definition after caveat, got %d", len(schema.Definitions))
	}
	if schema.Definitions[0].Name != "user" {
		t.Errorf("expected definition 'user', got %q", schema.Definitions[0].Name)
	}
}

func TestParseSkipCaveatWithNestedBraces(t *testing.T) {
	// Construct tokens manually since the lexer filters caveats
	tokens := []zedlexer.Token{
		tok(zedlexer.CAVEAT, "caveat"),
		tok(zedlexer.IDENTIFIER, "name"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RBRACE, "}"),
		tok(zedlexer.RBRACE, "}"),
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "user"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RBRACE, "}"),
	}
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(schema.Definitions))
	}
}

func TestParseDefinitionMissingName(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.LBRACE, "{"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing definition name")
	}
}

func TestParseDefinitionMissingLBrace(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "user"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing opening brace")
	}
}

func TestParseDefinitionInvalidMember(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "user"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.IDENTIFIER, "invalid"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid member in definition")
	}
}

func TestParseRelationMissingName(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RELATION, "relation"),
		tok(zedlexer.COLON, ":"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing relation name")
	}
}

func TestParseRelationMissingColon(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RELATION, "relation"),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.IDENTIFIER, "user"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing colon in relation")
	}
}

func TestParseRelationMissingSubjectAfterPipe(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RELATION, "relation"),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.COLON, ":"),
		tok(zedlexer.IDENTIFIER, "user"),
		tok(zedlexer.OR, "|"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing subject after pipe")
	}
}

func TestParseRelationMissingSubjectType(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.RELATION, "relation"),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.COLON, ":"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing subject type")
	}
}

func TestParsePermissionMissingName(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.EQUAL, "="),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing permission name")
	}
}

func TestParsePermissionMissingEquals(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.IDENTIFIER, "owner"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing equals in permission")
	}
}

func TestParsePermissionInvalidExpression(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid permission expression")
	}
}

func TestParseExclusionErrorInRightSide(t *testing.T) {
	// permission view = owner - (missing right side)
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.MINUS, "-"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid right side of exclusion")
	}
}

func TestParseIntersectionErrorInRightSide(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.AND, "&"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid right side of intersection")
	}
}

func TestParseUnionErrorInRightSide(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.PLUS, "+"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid right side of union")
	}
}

func TestParseArrowMissingIdentifier(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.IDENTIFIER, "parent"),
		tok(zedlexer.ARROW, "->"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing identifier after arrow")
	}
}

func TestParseArrowNonRelationRefLeft(t *testing.T) {
	// (a + b) -> view — left side is not a RelationRef
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.LBRACKETS, "("),
		tok(zedlexer.IDENTIFIER, "a"),
		tok(zedlexer.PLUS, "+"),
		tok(zedlexer.IDENTIFIER, "b"),
		tok(zedlexer.RBRACKETS, ")"),
		tok(zedlexer.ARROW, "->"),
		tok(zedlexer.IDENTIFIER, "perm"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for non-RelationRef left side of arrow")
	}
}

func TestParseGroupedExpressionMissingCloseParen(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.LBRACKETS, "("),
		tok(zedlexer.IDENTIFIER, "owner"),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for missing closing paren")
	}
}

func TestParseGroupedExpressionErrorInside(t *testing.T) {
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
		tok(zedlexer.LBRACKETS, "("),
		tok(zedlexer.RBRACE, "}"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for invalid expression inside parens")
	}
}

func TestParsePrimaryAtEnd(t *testing.T) {
	// Permission expression that ends abruptly
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
		tok(zedlexer.IDENTIFIER, "doc"),
		tok(zedlexer.LBRACE, "{"),
		tok(zedlexer.PERMISSION, "permission"),
		tok(zedlexer.IDENTIFIER, "view"),
		tok(zedlexer.EQUAL, "="),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error when expression is at end of input")
	}
}

func TestPeekAtEnd(t *testing.T) {
	// Parse empty token slice — tests peek returning EOF
	schema, err := Parse(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 0 {
		t.Errorf("expected 0 definitions for nil tokens")
	}
}

func TestParseDefinitionErrorPropagation(t *testing.T) {
	// definition keyword followed by end of input
	tokens := []zedlexer.Token{
		tok(zedlexer.DEFINITION, "definition"),
	}
	_, err := Parse(tokens)
	if err == nil {
		t.Fatal("expected error for incomplete definition")
	}
}

func TestParseSkipCaveatAtEnd(t *testing.T) {
	// Caveat without any braces — hits isAtEnd before finding opening brace
	tokens := []zedlexer.Token{
		tok(zedlexer.CAVEAT, "caveat"),
		tok(zedlexer.IDENTIFIER, "name"),
	}
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(schema.Definitions))
	}
}

// Tests below exercise internal parser methods directly to cover
// defensive error paths that are unreachable via the public API.

func TestInternalSkipCaveatExpectFails(t *testing.T) {
	// Call skipCaveat when next token is not CAVEAT
	p := &parser{tokens: []zedlexer.Token{tok(zedlexer.IDENTIFIER, "x")}, pos: 0}
	err := p.skipCaveat()
	if err == nil {
		t.Fatal("expected error from skipCaveat when token is not CAVEAT")
	}
}

func TestInternalParseDefinitionExpectFails(t *testing.T) {
	// Call parseDefinition when next token is not DEFINITION
	p := &parser{tokens: []zedlexer.Token{tok(zedlexer.IDENTIFIER, "x")}, pos: 0}
	_, err := p.parseDefinition()
	if err == nil {
		t.Fatal("expected error from parseDefinition when token is not DEFINITION")
	}
}

func TestInternalParseRelationExpectFails(t *testing.T) {
	// Call parseRelation when next token is not RELATION
	p := &parser{tokens: []zedlexer.Token{tok(zedlexer.IDENTIFIER, "x")}, pos: 0}
	_, err := p.parseRelation()
	if err == nil {
		t.Fatal("expected error from parseRelation when token is not RELATION")
	}
}

func TestInternalParsePermissionExpectFails(t *testing.T) {
	// Call parsePermission when next token is not PERMISSION
	p := &parser{tokens: []zedlexer.Token{tok(zedlexer.IDENTIFIER, "x")}, pos: 0}
	_, err := p.parsePermission()
	if err == nil {
		t.Fatal("expected error from parsePermission when token is not PERMISSION")
	}
}

func TestInternalErrorfAtPrevPosZero(t *testing.T) {
	// errorfAtPrev with pos=0 — no previous token
	p := &parser{tokens: []zedlexer.Token{tok(zedlexer.IDENTIFIER, "x")}, pos: 0}
	pe := p.errorfAtPrev("test %s", "msg")
	if pe.Line != 0 || pe.Column != 0 {
		t.Errorf("expected line=0 col=0 for pos=0, got line=%d col=%d", pe.Line, pe.Column)
	}
	if pe.Message != "test msg" {
		t.Errorf("expected message 'test msg', got %q", pe.Message)
	}
}

func TestParseSchemaSkipCaveatErrorPropagation(t *testing.T) {
	// Caveat token followed by EOF — skipCaveat succeeds but caveat is incomplete
	// This tests that the schema loop handles caveat correctly even at end
	tokens := []zedlexer.Token{
		tok(zedlexer.CAVEAT, "caveat"),
	}
	schema, err := Parse(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Definitions) != 0 {
		t.Errorf("expected 0 definitions, got %d", len(schema.Definitions))
	}
}
