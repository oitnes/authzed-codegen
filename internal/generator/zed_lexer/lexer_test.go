package zedlexer

import (
	"testing"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "trailing whitespace produces EOF token",
			input: "definition  ",
			want: []Token{
				{DEFINITION, "definition", 1, 1},
				{EOF, "", 1, 13},
			},
		},
		{
			name:  "single identifier",
			input: "identifier",
			want: []Token{
				{IDENTIFIER, "identifier", 1, 1},
			},
		},
		{
			name:  "keywords",
			input: "definition relation permission",
			want: []Token{
				{DEFINITION, "definition", 1, 1},
				{RELATION, "relation", 1, 12},
				{PERMISSION, "permission", 1, 21},
			},
		},
		{
			name:  "caveat keyword",
			input: "caveat",
			want: []Token{
				{CAVEAT, "caveat", 1, 1},
			},
		},
		{
			name:  "symbols",
			input: "{ } : | + - = -> :* ( ) &",
			want: []Token{
				{LBRACE, "{", 1, 1},
				{RBRACE, "}", 1, 3},
				{COLON, ":", 1, 5},
				{OR, "|", 1, 7},
				{PLUS, "+", 1, 9},
				{MINUS, "-", 1, 11},
				{EQUAL, "=", 1, 13},
				{ARROW, "->", 1, 15},
				{WILDCARD, ":*", 1, 18},
				{LBRACKETS, "(", 1, 21},
				{RBRACKETS, ")", 1, 23},
				{AND, "&", 1, 25},
			},
		},
		{
			name:  "line comment filtered",
			input: "// comment\nidentifier",
			want: []Token{
				{IDENTIFIER, "identifier", 2, 1},
			},
		},
		{
			name:  "block comment filtered",
			input: "/* block comment */\nidentifier",
			want: []Token{
				{IDENTIFIER, "identifier", 2, 1},
			},
		},
		{
			name:  "block comment inline",
			input: "definition /* inline */ user",
			want: []Token{
				{DEFINITION, "definition", 1, 1},
				{IDENTIFIER, "user", 1, 25},
			},
		},
		{
			name:  "multiple lines",
			input: "definition name\n{\nrelation user\n}",
			want: []Token{
				{DEFINITION, "definition", 1, 1},
				{IDENTIFIER, "name", 1, 12},
				{LBRACE, "{", 2, 1},
				{RELATION, "relation", 3, 1},
				{IDENTIFIER, "user", 3, 10},
				{RBRACE, "}", 4, 1},
			},
		},
		{
			name:  "namespaced identifier",
			input: "bookingsvc/user",
			want: []Token{
				{IDENTIFIER, "bookingsvc/user", 1, 1},
			},
		},
		{
			name:  "complex expression",
			input: "permission view = self + admin->manage",
			want: []Token{
				{PERMISSION, "permission", 1, 1},
				{IDENTIFIER, "view", 1, 12},
				{EQUAL, "=", 1, 17},
				{IDENTIFIER, "self", 1, 19},
				{PLUS, "+", 1, 24},
				{IDENTIFIER, "admin", 1, 26},
				{ARROW, "->", 1, 31},
				{IDENTIFIER, "manage", 1, 33},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex() error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("Lex() returned %d tokens, want %d\ngot:  %v", len(got), len(tt.want), got)
			}

			for i, token := range got {
				if token != tt.want[i] {
					t.Errorf("token[%d] = %+v, want %+v", i, token, tt.want[i])
				}
			}
		})
	}
}

func TestLexIllegalTokenReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"at sign", "@"},
		{"hash", "#"},
		{"exclamation", "!"},
		{"bare slash", "/"},
		{"illegal in context", "definition user { @ }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Lex(tt.input)
			if err == nil {
				t.Error("expected error for illegal token, got nil")
			}
		})
	}
}

func TestLexLineColumnTracking(t *testing.T) {
	input := "definition\n  relation\n    permission"
	got, err := Lex(input)
	if err != nil {
		t.Fatalf("Lex() error: %v", err)
	}

	expected := []struct {
		line, col int
	}{
		{1, 1}, // definition
		{2, 3}, // relation
		{3, 5}, // permission
	}

	if len(got) != len(expected) {
		t.Fatalf("got %d tokens, want %d", len(got), len(expected))
	}

	for i, e := range expected {
		if got[i].Line != e.line || got[i].Column != e.col {
			t.Errorf("token[%d] at line %d col %d, want line %d col %d",
				i, got[i].Line, got[i].Column, e.line, e.col)
		}
	}
}

func TestFilterComments(t *testing.T) {
	tokens := []Token{
		{DEFINITION, "definition", 1, 1},
		{COMMENT, "//", 1, 12},
		{IDENTIFIER, "user", 2, 1},
		{COMMENT, "/*", 2, 6},
	}

	got := filterComments(tokens)

	if len(got) != 2 {
		t.Fatalf("filterComments returned %d tokens, want 2", len(got))
	}
	if got[0].Type != DEFINITION {
		t.Errorf("token[0].Type = %v, want DEFINITION", got[0].Type)
	}
	if got[1].Type != IDENTIFIER {
		t.Errorf("token[1].Type = %v, want IDENTIFIER", got[1].Type)
	}
}

func TestFilterCaveats(t *testing.T) {
	tokens := []Token{
		{DEFINITION, "definition", 1, 1},
		{CAVEAT, "caveat", 2, 1},
	}

	got := filterCaveats(tokens)

	// Current implementation is a passthrough (TODO)
	if len(got) != len(tokens) {
		t.Errorf("filterCaveats returned %d tokens, want %d", len(got), len(tokens))
	}
}

func TestHaveIllegal(t *testing.T) {
	t.Run("no illegal tokens", func(t *testing.T) {
		tokens := []Token{
			{DEFINITION, "definition", 1, 1},
			{IDENTIFIER, "user", 1, 12},
		}
		have, _ := haveIlligal(tokens)
		if have {
			t.Error("expected no illegal tokens")
		}
	})

	t.Run("has illegal token", func(t *testing.T) {
		tokens := []Token{
			{DEFINITION, "definition", 1, 1},
			{ILLEGAL, "@", 1, 12},
		}
		have, tok := haveIlligal(tokens)
		if !have {
			t.Error("expected illegal token to be found")
		}
		if tok.Literal != "@" {
			t.Errorf("expected illegal literal '@', got %q", tok.Literal)
		}
	})

	t.Run("empty tokens", func(t *testing.T) {
		have, _ := haveIlligal(nil)
		if have {
			t.Error("expected no illegal tokens for nil input")
		}
	})
}

func TestLexBlockCommentMultiline(t *testing.T) {
	input := "definition /* multi\nline\ncomment */ user"
	got, err := Lex(input)
	if err != nil {
		t.Fatalf("Lex() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d tokens, want 2\ntokens: %v", len(got), got)
	}
	if got[0].Type != DEFINITION {
		t.Errorf("token[0].Type = %v, want DEFINITION", got[0].Type)
	}
	if got[1].Type != IDENTIFIER || got[1].Literal != "user" {
		t.Errorf("token[1] = %+v, want IDENTIFIER 'user'", got[1])
	}
}

func TestLexBlockCommentAtEndOfInput(t *testing.T) {
	_, err := Lex("definition /* unclosed")
	if err == nil {
		t.Fatal("expected error for unterminated block comment")
	}
}
