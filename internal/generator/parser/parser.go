package parser

import (
	"fmt"

	"github.com/oitnes/authzed-codegen/internal/generator/ast"
	zedlexer "github.com/oitnes/authzed-codegen/internal/generator/zed_lexer"
)

// ParseError represents an error encountered during parsing with location info.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

type parser struct {
	tokens []zedlexer.Token
	pos    int
}

// Parse converts a slice of tokens into an AST Schema.
func Parse(tokens []zedlexer.Token) (*ast.Schema, error) {
	p := &parser{tokens: tokens, pos: 0}
	return p.parseSchema()
}

func (p *parser) parseSchema() (*ast.Schema, error) {
	schema := &ast.Schema{}

	for !p.isAtEnd() {
		switch p.peek().Type {
		case zedlexer.DEFINITION:
			def, err := p.parseDefinition()
			if err != nil {
				return nil, err
			}
			schema.Definitions = append(schema.Definitions, def)
		case zedlexer.CAVEAT:
			if err := p.skipCaveat(); err != nil {
				return nil, err
			}
		case zedlexer.EOF:
			return schema, nil
		default:
			return nil, p.errorf("expected 'definition' or 'caveat', got %q", p.peek().Literal)
		}
	}

	return schema, nil
}

func (p *parser) parseDefinition() (*ast.Definition, error) {
	if _, err := p.expect(zedlexer.DEFINITION); err != nil {
		return nil, err
	}

	nameToken, err := p.expect(zedlexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(zedlexer.LBRACE); err != nil {
		return nil, err
	}

	def := &ast.Definition{Name: nameToken.Literal}

	for !p.isAtEnd() && p.peek().Type != zedlexer.RBRACE {
		switch p.peek().Type {
		case zedlexer.RELATION:
			rel, err := p.parseRelation()
			if err != nil {
				return nil, err
			}
			def.Relations = append(def.Relations, rel)
		case zedlexer.PERMISSION:
			perm, err := p.parsePermission()
			if err != nil {
				return nil, err
			}
			def.Permissions = append(def.Permissions, perm)
		default:
			return nil, p.errorf("expected 'relation' or 'permission', got %q", p.peek().Literal)
		}
	}

	if _, err := p.expect(zedlexer.RBRACE); err != nil {
		return nil, err
	}

	return def, nil
}

func (p *parser) parseRelation() (*ast.Relation, error) {
	if _, err := p.expect(zedlexer.RELATION); err != nil {
		return nil, err
	}

	nameToken, err := p.expect(zedlexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(zedlexer.COLON); err != nil {
		return nil, err
	}

	rel := &ast.Relation{Name: nameToken.Literal}

	subjectType, err := p.parseSubjectType()
	if err != nil {
		return nil, err
	}
	rel.SubjectTypes = append(rel.SubjectTypes, subjectType)

	for !p.isAtEnd() && p.peek().Type == zedlexer.OR {
		p.advance()
		subjectType, err := p.parseSubjectType()
		if err != nil {
			return nil, err
		}
		rel.SubjectTypes = append(rel.SubjectTypes, subjectType)
	}

	return rel, nil
}

func (p *parser) parseSubjectType() (*ast.SubjectType, error) {
	typeToken, err := p.expect(zedlexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	st := &ast.SubjectType{TypeName: typeToken.Literal}

	if !p.isAtEnd() && p.peek().Type == zedlexer.WILDCARD {
		p.advance()
		st.IsWildcard = true
	}

	return st, nil
}

func (p *parser) parsePermission() (*ast.Permission, error) {
	if _, err := p.expect(zedlexer.PERMISSION); err != nil {
		return nil, err
	}

	nameToken, err := p.expect(zedlexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(zedlexer.EQUAL); err != nil {
		return nil, err
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.Permission{Name: nameToken.Literal, Expression: expr}, nil
}

// Expression parsing with operator precedence (lowest to highest):
// 1. Exclusion (-)
// 2. Intersection (&)
// 3. Union (+)
// 4. Arrow (->)
// 5. Primary (identifier or grouped expression)

func (p *parser) parseExpression() (ast.Expr, error) {
	return p.parseExclusion()
}

func (p *parser) parseExclusion() (ast.Expr, error) {
	left, err := p.parseIntersection()
	if err != nil {
		return nil, err
	}

	for !p.isAtEnd() && p.peek().Type == zedlexer.MINUS {
		p.advance()
		right, err := p.parseIntersection()
		if err != nil {
			return nil, err
		}
		left = &ast.ExclusionExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseIntersection() (ast.Expr, error) {
	left, err := p.parseUnion()
	if err != nil {
		return nil, err
	}

	for !p.isAtEnd() && p.peek().Type == zedlexer.AND {
		p.advance()
		right, err := p.parseUnion()
		if err != nil {
			return nil, err
		}
		left = &ast.IntersectionExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseUnion() (ast.Expr, error) {
	left, err := p.parseArrow()
	if err != nil {
		return nil, err
	}

	for !p.isAtEnd() && p.peek().Type == zedlexer.PLUS {
		p.advance()
		right, err := p.parseArrow()
		if err != nil {
			return nil, err
		}
		left = &ast.UnionExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseArrow() (ast.Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if !p.isAtEnd() && p.peek().Type == zedlexer.ARROW {
		p.advance()

		permToken, err := p.expect(zedlexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}

		relRef, ok := left.(*ast.RelationRef)
		if !ok {
			return nil, p.errorfAtPrev("arrow operator requires a relation reference on the left side")
		}

		return &ast.ArrowExpr{
			Relation:   relRef.Name,
			Permission: permToken.Literal,
		}, nil
	}

	return left, nil
}

func (p *parser) parsePrimary() (ast.Expr, error) {
	if !p.isAtEnd() && p.peek().Type == zedlexer.LBRACKETS {
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(zedlexer.RBRACKETS); err != nil {
			return nil, err
		}
		return expr, nil
	}

	if !p.isAtEnd() && p.peek().Type == zedlexer.IDENTIFIER {
		token := p.advance()
		return &ast.RelationRef{Name: token.Literal}, nil
	}

	return nil, p.errorf("expected identifier or '(' in expression")
}

// skipCaveat skips a caveat block by brace-matching.
func (p *parser) skipCaveat() error {
	if _, err := p.expect(zedlexer.CAVEAT); err != nil {
		return err
	}

	// Skip tokens until we find the opening brace
	depth := 0
	for !p.isAtEnd() {
		if p.peek().Type == zedlexer.LBRACE {
			depth++
			p.advance()
			break
		}
		p.advance()
	}

	// Skip until matching closing brace
	for !p.isAtEnd() && depth > 0 {
		switch p.peek().Type {
		case zedlexer.LBRACE:
			depth++
		case zedlexer.RBRACE:
			depth--
		}
		p.advance()
	}

	return nil
}

// Helper methods

func (p *parser) peek() zedlexer.Token {
	if p.isAtEnd() {
		return zedlexer.Token{Type: zedlexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() zedlexer.Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.tokens[p.pos-1]
}

func (p *parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == zedlexer.EOF
}

func (p *parser) expect(tokenType zedlexer.TokenType) (zedlexer.Token, error) {
	if p.isAtEnd() {
		return zedlexer.Token{}, &ParseError{
			Message: fmt.Sprintf("unexpected end of input, expected token type %v", tokenType),
		}
	}

	token := p.peek()
	if token.Type != tokenType {
		return zedlexer.Token{}, &ParseError{
			Line:    token.Line,
			Column:  token.Column,
			Message: fmt.Sprintf("expected token type %v, got %v (%q)", tokenType, token.Type, token.Literal),
		}
	}

	return p.advance(), nil
}

func (p *parser) errorf(format string, args ...any) *ParseError {
	token := p.peek()
	return &ParseError{
		Line:    token.Line,
		Column:  token.Column,
		Message: fmt.Sprintf(format, args...),
	}
}

func (p *parser) errorfAtPrev(format string, args ...any) *ParseError {
	if p.pos == 0 {
		return &ParseError{Message: fmt.Sprintf(format, args...)}
	}
	token := p.tokens[p.pos-1]
	return &ParseError{
		Line:    token.Line,
		Column:  token.Column,
		Message: fmt.Sprintf(format, args...),
	}
}
