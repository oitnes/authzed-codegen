package ast

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	IDENTIFIER
	SLASH
	LBRACE
	RBRACE
	COLON
	PIPE
	PLUS
	EQUAL
	MINUS_ARROW
	DEFINITION
	RELATION
	PERMISSION
	COMMENT
	WILDCARD
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, line: 1, column: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *Lexer) next() rune {
	char := l.peek()
	l.pos++
	if char == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return char
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.next()
	}
}

func (l *Lexer) Lex() []Token {
	var tokens []Token
	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}
		char := l.peek()
		line, column := l.line, l.column
		switch char {
		case '/':
			tokens = append(tokens, Token{SLASH, "/", line, column})
			l.next()
		case '{':
			tokens = append(tokens, Token{LBRACE, "{", line, column})
			l.next()
		case '}':
			tokens = append(tokens, Token{RBRACE, "}", line, column})
			l.next()
		case ':':
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
				tokens = append(tokens, Token{WILDCARD, ":*", line, column})
				l.pos += 2
				l.column += 2
			} else {
				tokens = append(tokens, Token{COLON, ":", line, column})
			}
			l.next()
		case '|':
			tokens = append(tokens, Token{PIPE, "|", line, column})
			l.next()
		case '+':
			tokens = append(tokens, Token{PLUS, "+", line, column})
			l.next()
		case '=':
			tokens = append(tokens, Token{EQUAL, "=", line, column})
			l.next()
		case '-':
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
				tokens = append(tokens, Token{MINUS_ARROW, "->", line, column})
				l.pos += 2
				l.column += 2
			} else {
				tokens = append(tokens, Token{ILLEGAL, string(char), line, column})
				l.next()
			}
		default:
			if unicode.IsLetter(char) || char == '_' {
				literal := l.readIdentifier()
				tokenType := IDENTIFIER
				switch literal {
				case "definition":
					tokenType = DEFINITION
				case "relation":
					tokenType = RELATION
				case "permission":
					tokenType = PERMISSION
				}
				tokens = append(tokens, Token{tokenType, literal, line, column})
			} else {
				tokens = append(tokens, Token{ILLEGAL, string(char), line, column})
				l.next()
			}
		}
	}
	tokens = append(tokens, Token{EOF, "", l.line, l.column})
	return tokens
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_' {
		l.next()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) Error(msg string) error {
	return fmt.Errorf("lexer error: %s at line %d, column %d", msg, l.line, l.column)
}
