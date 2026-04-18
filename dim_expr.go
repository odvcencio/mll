package mll

import (
	"fmt"
	"strings"
)

// DimOp is a dimension expression operator.
type DimOp uint8

const (
	DimOpAdd DimOp = 0
	DimOpSub DimOp = 1
	DimOpMul DimOp = 2
	DimOpDiv DimOp = 3
)

// DimExpr is a binary expression tree over dimensions.
// Left and Right are operand dimensions (which may themselves be expressions).
type DimExpr struct {
	Op    DimOp
	Left  Dimension
	Right Dimension
}

// NewDimExpr constructs an expression with two leaf-style operands.
func NewDimExpr(op DimOp, left, right Dimension) *DimExpr {
	return &DimExpr{Op: op, Left: left, Right: right}
}

// Depth returns the maximum depth of the expression tree rooted here.
// A leaf (literal or symbol) has depth 1.
func (e *DimExpr) Depth() int {
	if e == nil {
		return 0
	}
	return 1 + maxInt(dimDepth(e.Left), dimDepth(e.Right))
}

func dimDepth(d Dimension) int {
	if d.Kind == DimKindExpr {
		return d.Expr.Depth()
	}
	return 1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// String returns the canonical text form of the expression.
func (e *DimExpr) String() string {
	if e == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(e.Left.String())
	switch e.Op {
	case DimOpAdd:
		sb.WriteString(" + ")
	case DimOpSub:
		sb.WriteString(" - ")
	case DimOpMul:
		sb.WriteString(" * ")
	case DimOpDiv:
		sb.WriteString(" / ")
	}
	sb.WriteString(e.Right.String())
	return sb.String()
}

// Equal reports deep equality of two expressions.
func (e *DimExpr) Equal(other *DimExpr) bool {
	if e == nil || other == nil {
		return e == other
	}
	if e.Op != other.Op {
		return false
	}
	return dimEqual(e.Left, other.Left) && dimEqual(e.Right, other.Right)
}

func dimEqual(a, b Dimension) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case DimKindLiteral:
		return a.Value == b.Value
	case DimKindSymbol:
		return a.Symbol == b.Symbol
	case DimKindExpr:
		return a.Expr.Equal(b.Expr)
	}
	return false
}

// NormalizeDim applies canonical normalization and returns the normalized
// dimension. Returns a new Dimension — crucially, if an inner expression
// folds to a pure literal, the parent sees it as a literal Kind, which is
// required for constant folding to propagate up the tree.
//
// The stringIndex argument maps symbol names to STRG indices for commutative
// reordering; pass nil for lexicographic symbol ordering.
func NormalizeDim(d Dimension, stringIndex map[string]uint32) Dimension {
	switch d.Kind {
	case DimKindLiteral, DimKindSymbol:
		return d
	case DimKindExpr:
		if d.Expr == nil {
			return d
		}
		// Recurse into children first (post-order).
		left := NormalizeDim(d.Expr.Left, stringIndex)
		right := NormalizeDim(d.Expr.Right, stringIndex)
		// Constant folding: if both operands are literals after recursion, collapse
		// the whole expression to a literal Dimension.
		if left.Kind == DimKindLiteral && right.Kind == DimKindLiteral {
			folded, ok := foldLiterals(d.Expr.Op, left.Value, right.Value)
			if ok {
				return DimLiteral(folded)
			}
		}
		// Commutative reordering for add and multiply.
		op := d.Expr.Op
		if op == DimOpAdd || op == DimOpMul {
			if dimOrder(right, left, stringIndex) < 0 {
				left, right = right, left
			}
		}
		return Dimension{
			Kind: DimKindExpr,
			Expr: &DimExpr{Op: op, Left: left, Right: right},
		}
	default:
		return d
	}
}

func foldLiterals(op DimOp, a, b int64) (int64, bool) {
	switch op {
	case DimOpAdd:
		return a + b, true
	case DimOpSub:
		return a - b, true
	case DimOpMul:
		return a * b, true
	case DimOpDiv:
		if b == 0 {
			return 0, false
		}
		return a / b, true
	}
	return 0, false
}

// dimOrder returns < 0, 0, or > 0 comparing two dims for canonical ordering.
//
// Ordering rule: symbols come before literals in the canonical form for both
// add and multiply. The test vector TestDimNormalizeCommutative pins the
// behavior end to end.
//
// Within the same kind:
//   - Symbols sort by string-table index ascending (falling back to
//     lexicographic comparison when no string index is available).
//   - Literals sort by integer value ascending, using explicit branches to
//     avoid int64 subtraction overflow.
func dimOrder(a, b Dimension, stringIndex map[string]uint32) int {
	aIsSym := a.Kind == DimKindSymbol
	bIsSym := b.Kind == DimKindSymbol
	if aIsSym && !bIsSym {
		return -1
	}
	if !aIsSym && bIsSym {
		return 1
	}
	if aIsSym && bIsSym {
		if stringIndex != nil {
			return int(stringIndex[a.Symbol]) - int(stringIndex[b.Symbol])
		}
		return strings.Compare(a.Symbol, b.Symbol)
	}
	// Both literals: use explicit branches to avoid int64 overflow.
	if a.Kind == DimKindLiteral && b.Kind == DimKindLiteral {
		switch {
		case a.Value < b.Value:
			return -1
		case a.Value > b.Value:
			return 1
		default:
			return 0
		}
	}
	return 0
}

// ValidateDimDepth reports an error if the dimension's expression tree
// exceeds the v1.0 depth limit of 8.
func ValidateDimDepth(d Dimension) error {
	depth := dimDepth(d)
	if depth > 8 {
		return fmt.Errorf("mll: dimension depth %d exceeds limit 8", depth)
	}
	return nil
}
