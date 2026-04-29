package ast

import "testing"

func TestExprNodeMarkers(t *testing.T) {
	// Verify all expression types implement the Expr interface.
	var exprs []Expr
	exprs = append(exprs, &UnionExpr{})
	exprs = append(exprs, &IntersectionExpr{})
	exprs = append(exprs, &ExclusionExpr{})
	exprs = append(exprs, &ArrowExpr{})
	exprs = append(exprs, &RelationRef{})

	for _, e := range exprs {
		e.exprNode() // call marker method
	}

	if len(exprs) != 5 {
		t.Errorf("expected 5 expression types, got %d", len(exprs))
	}
}
