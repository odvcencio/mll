package mll

import (
	"bytes"
	"testing"
)

func TestTypeSectionTensorRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewTypeBuilder()
	dIdx := strg.Intern("D")
	dDim := DimSymbol("D")
	dDim.SymbolIdx = dIdx
	idx := b.AddTensorType(strg.Intern("t_embed"), DTypeF32, []Dimension{dDim, DimLiteral(64)})
	if idx != 0 {
		t.Errorf("first type index: got %d", idx)
	}
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadTypeSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 1 {
		t.Fatalf("count: got %d", len(decoded.Decls))
	}
	if decoded.Decls[0].Kind != TypeKindTensor || decoded.Decls[0].DType != DTypeF32 {
		t.Fatalf("type: got %+v", decoded.Decls[0])
	}
	if len(decoded.Decls[0].Shape) != 2 {
		t.Fatalf("shape rank: got %d", len(decoded.Decls[0].Shape))
	}
}
