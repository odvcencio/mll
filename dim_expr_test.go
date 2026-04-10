package mll

import "testing"

func TestDimExprBasic(t *testing.T) {
	e := NewDimExpr(DimOpAdd, DimSymbol("D"), DimLiteral(4))
	if e.Depth() != 2 {
		t.Fatalf("depth: got %d", e.Depth())
	}
}

func TestDimNormalizeCommutative(t *testing.T) {
	// 4 + D should normalize to D + 4 (symbol before literal in add)
	d1 := Dimension{Kind: DimKindExpr, Expr: NewDimExpr(DimOpAdd, DimLiteral(4), DimSymbol("D"))}
	d2 := Dimension{Kind: DimKindExpr, Expr: NewDimExpr(DimOpAdd, DimSymbol("D"), DimLiteral(4))}
	n1 := NormalizeDim(d1, nil)
	n2 := NormalizeDim(d2, nil)
	if !dimEqual(n1, n2) {
		t.Fatalf("normalization failed: d1=%s d2=%s", n1, n2)
	}
}

func TestDimNormalizeConstFold(t *testing.T) {
	// D + (2 + 2) should fold to D + 4
	inner := Dimension{Kind: DimKindExpr, Expr: NewDimExpr(DimOpAdd, DimLiteral(2), DimLiteral(2))}
	outer := Dimension{Kind: DimKindExpr, Expr: NewDimExpr(DimOpAdd, DimSymbol("D"), inner)}
	normalized := NormalizeDim(outer, nil)
	// Expect: root is an expression whose right operand is a literal 4.
	if normalized.Kind != DimKindExpr {
		t.Fatalf("root kind: got %d", normalized.Kind)
	}
	if normalized.Expr.Right.Kind != DimKindLiteral || normalized.Expr.Right.Value != 4 {
		t.Fatalf("right: got %+v", normalized.Expr.Right)
	}
}

func TestDimNormalizePureLiteralCollapses(t *testing.T) {
	// (2 + 3) should collapse entirely to DimLiteral(5)
	d := Dimension{Kind: DimKindExpr, Expr: NewDimExpr(DimOpAdd, DimLiteral(2), DimLiteral(3))}
	normalized := NormalizeDim(d, nil)
	if normalized.Kind != DimKindLiteral || normalized.Value != 5 {
		t.Fatalf("pure literal fold: got %+v", normalized)
	}
}

func TestDimExprDepthLimit(t *testing.T) {
	// Depth 9 should be rejected
	e := DimSymbol("X")
	for i := 0; i < 9; i++ {
		e = Dimension{
			Kind: DimKindExpr,
			Expr: NewDimExpr(DimOpAdd, e, DimLiteral(1)),
		}
	}
	if err := ValidateDimDepth(e); err == nil {
		t.Fatal("depth 9 should be rejected")
	}
}
