package mll

import (
	"bytes"
	"testing"
)

func TestDimsSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewDimsBuilder()
	b.Add(DimDecl{NameIdx: strg.Intern("V"), Bound: DimBoundDynamic})
	b.Add(DimDecl{NameIdx: strg.Intern("D"), Bound: DimBoundStatic, Value: 384})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadDimsSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 2 {
		t.Fatalf("count: got %d", len(decoded.Decls))
	}
	if decoded.Decls[1].Bound != DimBoundStatic || decoded.Decls[1].Value != 384 {
		t.Fatalf("static dim: got %+v", decoded.Decls[1])
	}
}
