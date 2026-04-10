package mll

import (
	"fmt"
	"io"
)

// WriteDimension encodes a dimension to w using the canonical binary form.
// Format:
//
//	u8 kind
//	kind == literal:  varint i64 value
//	kind == symbol:   u32 string_table_idx
//	kind == expr:     u8 op + Dimension(left) + Dimension(right)
//
// The dimension is validated for depth before normalization, then normalized
// before encoding. This enforces the canonicalization rule that sealed/weights-only
// section bodies contain normalized dim expressions, without requiring every
// section builder to remember to call NormalizeDim. The function validates the
// input depth against the limit of 8.
func WriteDimension(w io.Writer, d Dimension) error {
	return WriteDimensionWithIndex(w, d, nil)
}

// WriteDimensionWithIndex is like WriteDimension but takes a string-table
// index map so symbol-based commutative reordering uses the interned indices
// rather than lexicographic symbol names. Callers who have finalized the
// string table pass their index map; callers who haven't pass nil.
func WriteDimensionWithIndex(w io.Writer, d Dimension, stringIndex map[string]uint32) error {
	if err := ValidateDimDepth(d); err != nil {
		return err
	}
	normalized := NormalizeDim(d, stringIndex)
	return writeDimensionInternal(w, normalized)
}

func writeDimensionInternal(w io.Writer, d Dimension) error {
	if _, err := w.Write([]byte{byte(d.Kind)}); err != nil {
		return err
	}
	switch d.Kind {
	case DimKindLiteral:
		return WriteVarint(w, d.Value)
	case DimKindSymbol:
		return WriteUint32LE(w, d.SymbolIdx)
	case DimKindExpr:
		if d.Expr == nil {
			return fmt.Errorf("mll: expr dim with nil expression")
		}
		if _, err := w.Write([]byte{byte(d.Expr.Op)}); err != nil {
			return err
		}
		if err := writeDimensionInternal(w, d.Expr.Left); err != nil {
			return err
		}
		return writeDimensionInternal(w, d.Expr.Right)
	default:
		return fmt.Errorf("mll: invalid dim kind %d", d.Kind)
	}
}

// ReadDimension decodes a dimension from b. Returns the dimension, the number
// of bytes consumed, and any error.
func ReadDimension(b []byte) (Dimension, int, error) {
	return readDimensionAt(b, 0, 0)
}

// readDimensionAt reads a dimension from b starting at offset; returns the
// dimension, the new offset (= bytes consumed from start of b), and any error.
func readDimensionAt(b []byte, offset, depth int) (Dimension, int, error) {
	if depth > 8 {
		return Dimension{}, 0, fmt.Errorf("mll: dimension depth exceeds 8")
	}
	if offset >= len(b) {
		return Dimension{}, 0, fmt.Errorf("mll: dimension truncated at offset %d", offset)
	}
	kind := DimKind(b[offset])
	offset++
	switch kind {
	case DimKindLiteral:
		v, n, err := ReadVarint(b[offset:])
		if err != nil {
			return Dimension{}, 0, err
		}
		return Dimension{Kind: DimKindLiteral, Value: v}, offset + n, nil
	case DimKindSymbol:
		if offset+4 > len(b) {
			return Dimension{}, 0, fmt.Errorf("mll: symbol dim truncated")
		}
		idx, err := ReadUint32LE(b[offset : offset+4])
		if err != nil {
			return Dimension{}, 0, err
		}
		return Dimension{Kind: DimKindSymbol, SymbolIdx: idx}, offset + 4, nil
	case DimKindExpr:
		if offset >= len(b) {
			return Dimension{}, 0, fmt.Errorf("mll: expr dim missing op")
		}
		op := DimOp(b[offset])
		offset++
		left, newOff, err := readDimensionAt(b, offset, depth+1)
		if err != nil {
			return Dimension{}, 0, err
		}
		right, newOff, err := readDimensionAt(b, newOff, depth+1)
		if err != nil {
			return Dimension{}, 0, err
		}
		expr := NewDimExpr(op, left, right)
		return Dimension{Kind: DimKindExpr, Expr: expr}, newOff, nil
	default:
		return Dimension{}, 0, fmt.Errorf("mll: invalid dim kind %d", kind)
	}
}

// WriteShape encodes a shape (sequence of dimensions) to w.
// Format: u32 rank + Dimension[rank].
func WriteShape(w io.Writer, shape []Dimension) error {
	if err := WriteUint32LE(w, uint32(len(shape))); err != nil {
		return err
	}
	for _, d := range shape {
		if err := WriteDimension(w, d); err != nil {
			return err
		}
	}
	return nil
}

// ReadShape decodes a shape from b. Returns the shape, bytes consumed, and any error.
func ReadShape(b []byte) ([]Dimension, int, error) {
	if len(b) < 4 {
		return nil, 0, fmt.Errorf("mll: shape needs at least 4 bytes")
	}
	rank, _ := ReadUint32LE(b[:4])
	cursor := 4
	shape := make([]Dimension, rank)
	for i := uint32(0); i < rank; i++ {
		d, n, err := ReadDimension(b[cursor:])
		if err != nil {
			return nil, 0, err
		}
		shape[i] = d
		cursor += n
	}
	return shape, cursor, nil
}
