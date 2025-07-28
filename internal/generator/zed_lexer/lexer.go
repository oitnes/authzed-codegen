package zed_lexer

import (
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF

	LBRACE
	RBRACE
	LBRACKETS
	RBRACKETS
	COLON
	OR
	AND
	PLUS
	MINUS
	EQUAL
	ARROW
	WILDCARD

	IDENTIFIER
	DEFINITION
	RELATION
	PERMISSION
	CAVEAT
	COMMENT
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	endChar    = 0
	endLine    = 10 //   \n
	slash      = 47 //   /
	underscore = 95 //   _
	star       = 42 //   *
)

type Lexer struct {
	InputCode string

	pos    int // cursor throw InputCode
	line   int // current line number
	column int // current column number
}

// Lex scans the InputCode and returns a slice of tokens, resetting line, column, and position counters before processing.
func (l *Lexer) Lex() []Token {
	l.line = 1
	l.column = 1
	l.pos = 0

	var tokens []Token

	for l.haveNext() {
		tokens = append(tokens, l.readToken())
	}

	return tokens
}

// readToken reads the next token from the InputCode and returns it as a Token. It handles various token types and skips whitespaces.
func (l *Lexer) readToken() Token {
	// skip whitespaces between tokens
	for unicode.IsSpace(l.peek()) {
		l.skip()
	}

	char := l.peek()

	// if skipped all, return as EOF
	if char == endChar {
		return Token{EOF, "", l.line, l.column}
	}

	line, column := l.line, l.column

	switch char {
	case slash:
		if l.peekForward() == slash {
			l.skipLineComment()
			return Token{COMMENT, "//", line, column}
		} else if l.peekForward() == star {
			l.skipComplexComment()
			return Token{COMMENT, "/*", line, column}
		} else {
			l.skip()
			return Token{ILLEGAL, "/", line, column}
		}
	case '{':
		l.skip()
		return Token{LBRACE, "{", line, column}
	case '}':
		l.skip()
		return Token{RBRACE, "}", line, column}
	case '(':
		l.skip()
		return Token{LBRACKETS, "(", line, column}
	case ')':
		l.skip()
		return Token{RBRACKETS, ")", line, column}
	case ':':
		if l.peekForward() == star {
			l.skipComplicatedSymbol(":*")
			return Token{WILDCARD, ":*", line, column}
		} else {
			l.skip()
			return Token{COLON, ":", line, column}
		}
	case '|':
		l.skip()
		return Token{OR, "|", line, column}
	case '&':
		l.skip()
		return Token{AND, "&", line, column}
	case '+':
		l.skip()
		return Token{PLUS, "+", line, column}
	case '=':
		l.skip()
		return Token{EQUAL, "=", line, column}
	case '-':
		if l.peekForward() == '>' {
			l.skipComplicatedSymbol("->")
			return Token{ARROW, "->", line, column}
		} else {
			l.skip()
			return Token{MINUS, string(char), line, column}
		}
	default:
		if l.isIdentifierPart(char) {
			literal := l.readIdentifier()
			tokenType := IDENTIFIER

			switch literal {
			case "caveat":
				tokenType = CAVEAT
			case "definition":
				tokenType = DEFINITION
			case "relation":
				tokenType = RELATION
			case "permission":
				tokenType = PERMISSION
			}

			return Token{tokenType, literal, line, column}
		} else {
			l.skip()
			return Token{ILLEGAL, string(char), line, column}
		}
	}
}

// readIdentifier read and return some word
func (l *Lexer) readIdentifier() string {
	start := l.pos
	for l.isIdentifierPart(l.peek()) {
		l.skip()
	}

	return l.InputCode[start:l.pos]
}

// isIdentifierPart return true if symbol part of some identifier or world
func (l *Lexer) isIdentifierPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == underscore || ch == slash
}

func (l *Lexer) isEndOfLine(ch rune) bool {
	return ch == endLine || ch == endChar
}

// shiftToNextLine advances the lexer to the next line, updating the line and column counters accordingly.
func (l *Lexer) skipLineComment() {
	for !l.isEndOfLine(l.peek()) {
		l.skip()
	}
}

func (l *Lexer) skipComplexComment() {
	for l.peek() != endChar {
		if l.peek() == star && l.peekForward() == slash {
			break
		}

		l.skip()
	}
}

// skipComplicatedSymbol move cursor for len of s positions
func (l *Lexer) skipComplicatedSymbol(s string) {
	for range s {
		l.skip()
	}
}

// shift move cursor to next symbol
func (l *Lexer) skip() {
	char := l.peek()
	l.pos++

	if l.isEndOfLine(char) {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

// peekForward return rune at next position
func (l *Lexer) peekForward() rune {
	if !l.haveNextN(2) {
		return endChar
	}

	return rune(l.InputCode[l.pos+1])
}

// peek return current char at pos or endChar
func (l *Lexer) peek() rune {
	if !l.haveNext() {
		return endChar
	}

	return rune(l.InputCode[l.pos])
}

// haveNext return true if symbol at current pos exists in InputCode
func (l *Lexer) haveNext() bool {
	return l.haveNextN(1)
}

// haveNextN return true if n symbol (include current pos) exists in InputCode
func (l *Lexer) haveNextN(n int) bool {
	return l.pos+n <= len(l.InputCode)
}
