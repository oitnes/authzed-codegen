package ast

// Schema represents the complete parsed Zed schema
type Schema struct {
	Definitions []*Definition
}

// Definition represents a single object type definition
type Definition struct {
	Name        string        // e.g., "public_forum" or "bookingsvc/booking"
	Relations   []*Relation   // Relations defined on this type
	Permissions []*Permission // Permissions computed on this type
}

// Relation represents a relation definition
type Relation struct {
	Name         string         // e.g., "owner", "member"
	SubjectTypes []*SubjectType // Types that can be subjects of this relation
}

// SubjectType represents a type that can be a subject in a relation
type SubjectType struct {
	TypeName   string // e.g., "user" or "bookingsvc/user"
	IsWildcard bool   // true for "user:*"
}

// Permission represents a permission definition
type Permission struct {
	Name       string // e.g., "view", "edit"
	Expression Expr   // Expression that computes this permission
}

// Expr is the interface for permission expressions
type Expr interface {
	exprNode()
}

// UnionExpr represents a union operation (OR): left + right
type UnionExpr struct {
	Left  Expr
	Right Expr
}

// IntersectionExpr represents an intersection operation (AND): left & right
type IntersectionExpr struct {
	Left  Expr
	Right Expr
}

// ExclusionExpr represents an exclusion operation: left - right
type ExclusionExpr struct {
	Left  Expr
	Right Expr
}

// ArrowExpr represents arrow traversal: relation->permission
type ArrowExpr struct {
	Relation   string // The relation to traverse
	Permission string // The permission to check on the related object
}

// RelationRef represents a reference to a relation by name
type RelationRef struct {
	Name string // The relation name
}

// Implement exprNode() for all expression types
func (*UnionExpr) exprNode()        {}
func (*IntersectionExpr) exprNode() {}
func (*ExclusionExpr) exprNode()    {}
func (*ArrowExpr) exprNode()        {}
func (*RelationRef) exprNode()      {}
