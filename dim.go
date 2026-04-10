package mll

import "strconv"

// DimKind identifies the form of a Dimension.
type DimKind uint8

const (
	DimKindLiteral DimKind = 0
	DimKindSymbol  DimKind = 1
	DimKindExpr    DimKind = 2
)

// Dimension is a first-class primitive representing a tensor dimension.
// It is either a literal integer, a symbolic reference to a named dim,
// or an expression tree over other dims.
type Dimension struct {
	Kind      DimKind
	Value     int64    // valid when Kind == DimKindLiteral
	Symbol    string   // valid when Kind == DimKindSymbol (Go-side name)
	SymbolIdx uint32   // valid when Kind == DimKindSymbol (string table index)
	Expr      *DimExpr // valid when Kind == DimKindExpr (defined in dim_expr.go)
}

// DimLiteral constructs a literal dimension.
func DimLiteral(v int64) Dimension {
	return Dimension{Kind: DimKindLiteral, Value: v}
}

// DimSymbol constructs a symbolic dimension reference.
func DimSymbol(name string) Dimension {
	return Dimension{Kind: DimKindSymbol, Symbol: name}
}

// String returns the canonical text form of the dimension.
func (d Dimension) String() string {
	switch d.Kind {
	case DimKindLiteral:
		return strconv.FormatInt(d.Value, 10)
	case DimKindSymbol:
		return d.Symbol
	case DimKindExpr:
		if d.Expr != nil {
			return d.Expr.String()
		}
		return "<empty-expr>"
	default:
		return "<invalid-dim>"
	}
}

// DimDecl is a module-level dimension declaration.
// Bound dims have a concrete value; free dims are bound at load/entry time.
type DimDecl struct {
	Name  string
	Bound bool  // true if Value is set
	Value int64 // valid when Bound is true
}

