package ast

import "fmt"

// Parser represents a parser for the language.
type Parser struct {
	tokens      []Token           // Slice of tokens to parse.
	current     int               // Current token index.
	Definitions []*DefinitionNode // Added to store multiple definitions
}

// NewParser creates a new Parser with the given tokens.
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, Definitions: []*DefinitionNode{}}
}

func (p *Parser) ParseDefinitions() ([]*DefinitionNode, error) {
	for p.peek().Type != EOF {
		def, err := p.parseDefinition()
		if err != nil {
			return nil, err
		}
		p.Definitions = append(p.Definitions, def)
	}
	return p.Definitions, nil
}

// peek returns the current token without advancing the parser.
func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

// consume advances the parser to the next token and checks if it matches the expected type.
func (p *Parser) consume(expected TokenType) Token {
	if p.peek().Type == expected {
		token := p.peek()
		p.current++
		return token
	}
	panic(fmt.Sprintf("Expected %v, got %v", expected, p.peek()))
}

// parseDefinition parses a definition statement.
func (p *Parser) parseDefinition() (*DefinitionNode, error) {
	p.consume(DEFINITION)                   // Consume "definition" keyword.
	prefix := p.consume(IDENTIFIER).Literal // Consume and store object type.
	p.consume(SLASH)                        // Consume slash.
	name := p.consume(IDENTIFIER).Literal   // Consume and store definition name.
	p.consume(LBRACE)                       // Consume left brace.

	def := &DefinitionNode{
		ObjectType: ObjectType{
			Name:   name,
			Prefix: prefix,
		},
		Relations:   []*RelationNode{},
		Permissions: []*PermissionNode{},
	}

	// Parse relations and permissions until right brace is encountered.
	for p.peek().Type == RELATION || p.peek().Type == PERMISSION {
		if p.peek().Type == RELATION {
			rel, err := p.parseRelation()
			if err != nil {
				return nil, err
			}
			def.Relations = append(def.Relations, rel)
		} else if p.peek().Type == PERMISSION {
			perm, err := p.parsePermission()
			if err != nil {
				return nil, err
			}
			def.Permissions = append(def.Permissions, perm)
		}
	}

	p.consume(RBRACE) // Consume right brace.
	return def, nil
}

// parsePermission parses a permission statement.
func (p *Parser) parsePermission() (*PermissionNode, error) {

	p.consume(PERMISSION)                 // Consume "permission" keyword.
	name := p.consume(IDENTIFIER).Literal // Consume and store permission name.
	p.consume(EQUAL)                      // Consume equal sign.
	expr, err := p.parsePermissionExpression()
	if err != nil {
		return nil, err
	}

	return &PermissionNode{Name: name, Expression: expr}, nil
}

// Parse starts the parsing process and returns the root DefinitionNode.
func (p *Parser) Parse() (*DefinitionNode, error) {
	return p.parseDefinition()
}

func (p *Parser) parsePermissionExpression() (PermissionExpressionNode, error) {
	return p.parseAdditiveExpression()
}

func (p *Parser) parseAdditiveExpression() (PermissionExpressionNode, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == PLUS {
		op := p.consume(PLUS).Literal
		right, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parsePrimaryExpression() (PermissionExpressionNode, error) {
	left, err := p.parseIdentifierChain()
	if err != nil {
		return nil, err
	}

	return left, nil
}

func (p *Parser) parseIdentifierChain() (PermissionExpressionNode, error) {
	var left PermissionExpressionNode
	left = &IdentifierNode{Value: p.consume(IDENTIFIER).Literal}

	for p.peek().Type == MINUS_ARROW {
		op := p.consume(MINUS_ARROW).Literal
		right := &IdentifierNode{Value: p.consume(IDENTIFIER).Literal}
		left = &BinaryOpNode{Operator: op, Left: left, Right: right} // Corrected: left is now ExpressionNode
	}

	return left, nil
}

func (p *Parser) parseRelation() (*RelationNode, error) {
	p.consume(RELATION)
	name := p.consume(IDENTIFIER).Literal
	p.consume(COLON)
	relationExpr, err := p.parseRelationExpression()
	if err != nil {
		return nil, err
	}
	return &RelationNode{Name: name, Expression: relationExpr}, nil
}

func (p *Parser) parseRelationExpression() (RelationExpressionNode, error) {
	left, err := p.parseSingleRelation()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == PIPE || p.peek().Type == WILDCARD {
		if p.peek().Type == WILDCARD {
			p.consume(WILDCARD)
		} else {
			p.consume(PIPE)
			right, err := p.parseSingleRelation()
			if err != nil {
				return nil, err
			}
			left = &UnionRelationNode{Left: left, Right: right}
		}
	}

	return left, nil
}

func (p *Parser) parseSingleRelation() (RelationExpressionNode, error) {
	relationType := ""
	for p.peek().Type == IDENTIFIER || p.peek().Type == SLASH {
		relationType += p.consume(p.peek().Type).Literal
	}
	return &SingleRelationNode{Value: relationType}, nil
}
