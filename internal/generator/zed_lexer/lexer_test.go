package zedlexer

import (
	"reflect"
	"testing"
)

func TestLexer_Lex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "empty InputCode",
			input: "",
			want:  []Token{},
		},
		{
			name:  "single identifier",
			input: "identifier",
			want: []Token{
				{Type: IDENTIFIER, Literal: "identifier", Line: 1, Column: 1},
			},
		},
		{
			name:  "special keywords",
			input: "definition relation permission",
			want: []Token{
				{Type: DEFINITION, Literal: "definition", Line: 1, Column: 1},
				{Type: RELATION, Literal: "relation", Line: 1, Column: 12},
				{Type: PERMISSION, Literal: "permission", Line: 1, Column: 21},
			},
		},
		{
			name:  "symbols",
			input: "/ { } : | + - = -> :* ( ) &",
			want: []Token{
				{Type: ILLEGAL, Literal: "/", Line: 1, Column: 1},
				{Type: LBRACE, Literal: "{", Line: 1, Column: 3},
				{Type: RBRACE, Literal: "}", Line: 1, Column: 5},
				{Type: COLON, Literal: ":", Line: 1, Column: 7},
				{Type: OR, Literal: "|", Line: 1, Column: 9},
				{Type: PLUS, Literal: "+", Line: 1, Column: 11},
				{Type: MINUS, Literal: "-", Line: 1, Column: 13},
				{Type: EQUAL, Literal: "=", Line: 1, Column: 15},
				{Type: ARROW, Literal: "->", Line: 1, Column: 17},
				{Type: WILDCARD, Literal: ":*", Line: 1, Column: 20},
				{Type: LBRACKETS, Literal: "(", Line: 1, Column: 23},
				{Type: RBRACKETS, Literal: ")", Line: 1, Column: 25},
				{Type: AND, Literal: "&", Line: 1, Column: 27},
			},
		},
		{
			name:  "comment",
			input: "// this is a comment\nidentifier",
			want: []Token{
				{Type: COMMENT, Literal: "//", Line: 1, Column: 1},
				{Type: IDENTIFIER, Literal: "identifier", Line: 2, Column: 1},
			},
		},
		{
			name:  "illegal character",
			input: "@",
			want: []Token{
				{Type: ILLEGAL, Literal: "@", Line: 1, Column: 1},
			},
		},
		{
			name:  "multiple lines",
			input: "definition name\n{\nrelation user\n}",
			want: []Token{
				{Type: DEFINITION, Literal: "definition", Line: 1, Column: 1},
				{Type: IDENTIFIER, Literal: "name", Line: 1, Column: 12},
				{Type: LBRACE, Literal: "{", Line: 2, Column: 1},
				{Type: RELATION, Literal: "relation", Line: 3, Column: 1},
				{Type: IDENTIFIER, Literal: "user", Line: 3, Column: 10},
				{Type: RBRACE, Literal: "}", Line: 4, Column: 1},
			},
		},
		{
			name:  "complex expression",
			input: "definition user {\n  permission view = self + admin\n  relation admin: group/admin\n}",
			want: []Token{
				{Type: DEFINITION, Literal: "definition", Line: 1, Column: 1},
				{Type: IDENTIFIER, Literal: "user", Line: 1, Column: 12},
				{Type: LBRACE, Literal: "{", Line: 1, Column: 17},
				{Type: PERMISSION, Literal: "permission", Line: 2, Column: 3},
				{Type: IDENTIFIER, Literal: "view", Line: 2, Column: 14},
				{Type: EQUAL, Literal: "=", Line: 2, Column: 19},
				{Type: IDENTIFIER, Literal: "self", Line: 2, Column: 21},
				{Type: PLUS, Literal: "+", Line: 2, Column: 26},
				{Type: IDENTIFIER, Literal: "admin", Line: 2, Column: 28},
				{Type: RELATION, Literal: "relation", Line: 3, Column: 3},
				{Type: IDENTIFIER, Literal: "admin", Line: 3, Column: 12},
				{Type: COLON, Literal: ":", Line: 3, Column: 17},
				{Type: IDENTIFIER, Literal: "group/admin", Line: 3, Column: 19},
				{Type: RBRACE, Literal: "}", Line: 4, Column: 1},
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{InputCode: tt.input}
			got := l.Lex()

			// If no tokens are expected and none are returned, the test passes
			if len(tt.want) == 0 && len(got) == 0 {
				return
			}

			// Check if the number of tokens matches
			if len(got) != len(tt.want) {
				t.Errorf("Lex() returned %d tokens, want %d tokens", len(got), len(tt.want))
				t.Errorf("Got tokens: %v", got)
				return
			}

			// Check each token
			for i, token := range got {
				if !reflect.DeepEqual(token, tt.want[i]) {
					t.Errorf("Lex()[%d] = %+v, want %+v", i, token, tt.want[i])
				}
			}
		})
	}
}

// TestLexer_LexWithLineColumnTracking tests the line and column tracking capability of the Lexer
func TestLexer_LexWithLineColumnTracking(t *testing.T) {
	input := "definition\n  relation\n    permission"
	l := &Lexer{InputCode: input}
	got := l.Lex()

	expected := []Token{
		{Type: DEFINITION, Literal: "definition", Line: 1, Column: 1},
		{Type: RELATION, Literal: "relation", Line: 2, Column: 3},
		{Type: PERMISSION, Literal: "permission", Line: 3, Column: 5},
	}

	if len(got) != len(expected) {
		t.Fatalf("Lex() returned %d tokens, want %d tokens", len(got), len(expected))
	}

	for i, token := range got {
		if token.Type != expected[i].Type ||
			token.Literal != expected[i].Literal ||
			token.Line != expected[i].Line ||
			token.Column != expected[i].Column {
			t.Errorf("Lex()[%d] = %+v, want %+v", i, token, expected[i])
		}
	}
}

// TestLexer_LexWithIllegalTokens tests the handling of illegal tokens
func TestLexer_LexWithIllegalTokens(t *testing.T) {
	input := "definition ! @ #"
	l := &Lexer{InputCode: input}
	got := l.Lex()

	expected := []Token{
		{Type: DEFINITION, Literal: "definition", Line: 1, Column: 1},
		{Type: ILLEGAL, Literal: "!", Line: 1, Column: 12},
		{Type: ILLEGAL, Literal: "@", Line: 1, Column: 14},
		{Type: ILLEGAL, Literal: "#", Line: 1, Column: 16},
	}

	if len(got) != len(expected) {
		t.Fatalf("Lex() returned %d tokens, want %d tokens", len(got), len(expected))
	}

	for i, token := range got {
		if token.Type != expected[i].Type ||
			token.Literal != expected[i].Literal ||
			token.Line != expected[i].Line ||
			token.Column != expected[i].Column {
			t.Errorf("Lex()[%d] = %+v, want %+v", i, token, expected[i])
		}
	}
}
