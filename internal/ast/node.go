package ast

import "fmt"

// Node represents a node in the Abstract Syntax Tree (AST).
type Node interface {
	String() string
}

type ObjectType struct {
	Name   string
	Prefix string
}

func (ot ObjectType) String() string {
	return fmt.Sprintf("%s/%s", ot.Prefix, ot.Name)
}

// DefinitionNode represents a definition of an object type with relations and permissions.
type DefinitionNode struct {
	ObjectType  ObjectType        // Type of the object (e.g., "init_booking/booking").
	Relations   []*RelationNode   // List of relations defined for the object.
	Permissions []*PermissionNode // List of permissions defined for the object.
}

// String returns a string representation of the DefinitionNode.
func (dn *DefinitionNode) String() string {
	str := fmt.Sprintf("definition %s/%s {\n", dn.ObjectType.Prefix, dn.ObjectType.Name)
	for _, rel := range dn.Relations {
		str += "    " + rel.String() + "\n"
	}
	for _, perm := range dn.Permissions {
		str += "    " + perm.String() + "\n"
	}
	str += "}\n"
	return str
}

// RelationNode represents a relation defined for an object.
type RelationNode struct {
	Name       string                 // Name of the relation (e.g., "owner").
	Expression RelationExpressionNode // Expression defining the relation (e.g., "init_booking/employee | init_booking/customer").
}

// String returns a string representation of the RelationNode.
func (rn *RelationNode) String() string {
	return fmt.Sprintf("relation %s: %s", rn.Name, rn.Expression)
}

// PermissionNode represents a permission defined for an object.
type PermissionNode struct {
	Name       string                   // Name of the permission (e.g., "write").
	Expression PermissionExpressionNode // Expression defining the permission (e.g., "creator + owner + owner->manage + creator->manage").
}

// String returns a string representation of the PermissionNode.
func (pn *PermissionNode) String() string {
	return fmt.Sprintf("permission %s = %s", pn.Name, pn.Expression)
}

// ExpressionNode represents an expression in the AST.

type PermissionExpressionNode interface {
	Node
}

type IdentifierNode struct {
	Value string
}

func (in *IdentifierNode) String() string {
	return in.Value
}

type BinaryOpNode struct {
	Operator string
	Left     PermissionExpressionNode
	Right    PermissionExpressionNode
}

func (bon *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", bon.Operator, bon.Left, bon.Right)
}

// RelationExpressionNode represents a relation expression in the AST.

type RelationExpressionNode interface {
	Node
}

type SingleRelationNode struct {
	Value string
}

func (srn *SingleRelationNode) String() string {
	return srn.Value
}

type UnionRelationNode struct {
	Left  RelationExpressionNode
	Right RelationExpressionNode
}

func (urn *UnionRelationNode) String() string {
	return fmt.Sprintf("(%s | %s)", urn.Left, urn.Right)
}
