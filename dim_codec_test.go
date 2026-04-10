package mll

import (
	"bytes"
	"testing"
)

func TestDimensionLiteralRoundTrip(t *testing.T) {
	d := DimLiteral(384)
	var buf bytes.Buffer
	if err := WriteDimension(&buf, d); err != nil {
		t.Fatal(err)
	}
	got, _, err := ReadDimension(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != DimKindLiteral || got.Value != 384 {
		t.Fatalf("got %+v", got)
	}
}

func TestDimensionSymbolRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	dIdx := strg.Intern("D")
	d := DimSymbol("D")
	d.SymbolIdx = dIdx
	var buf bytes.Buffer
	if err := WriteDimension(&buf, d); err != nil {
		t.Fatal(err)
	}
	got, _, err := ReadDimension(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != DimKindSymbol || got.SymbolIdx != dIdx {
		t.Fatalf("got %+v", got)
	}
}

func TestDimensionExprRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	dIdx := strg.Intern("D")
	left := DimSymbol("D")
	left.SymbolIdx = dIdx
	right := DimLiteral(4)
	d := Dimension{
		Kind: DimKindExpr,
		Expr: NewDimExpr(DimOpAdd, left, right),
	}
	var buf bytes.Buffer
	if err := WriteDimension(&buf, d); err != nil {
		t.Fatal(err)
	}
	got, _, err := ReadDimension(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != DimKindExpr {
		t.Fatalf("kind: got %d", got.Kind)
	}
	if got.Expr == nil || got.Expr.Op != DimOpAdd {
		t.Fatalf("expr: got %+v", got.Expr)
	}
}

func TestDimensionDepthEnforced(t *testing.T) {
	// Build a depth-9 expression.
	d := DimLiteral(1)
	for i := 0; i < 9; i++ {
		d = Dimension{
			Kind: DimKindExpr,
			Expr: NewDimExpr(DimOpAdd, d, DimLiteral(1)),
		}
	}
	var buf bytes.Buffer
	err := WriteDimension(&buf, d)
	if err == nil {
		t.Fatal("expected depth-limit error")
	}
}
