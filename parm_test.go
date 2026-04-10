package mll

import (
	"bytes"
	"testing"
)

func TestParmSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewParmBuilder()
	b.Add(ParmDecl{
		NameIdx:   strg.Intern("w"),
		TypeRef:   Ref{Tag: TagTYPE, Index: 0},
		Trainable: true,
	})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadParmSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 1 {
		t.Fatalf("count: got %d", len(decoded.Decls))
	}
	if !decoded.Decls[0].Trainable {
		t.Error("trainable lost")
	}
	if decoded.Decls[0].TypeRef.Tag != TagTYPE {
		t.Errorf("type ref: got %+v", decoded.Decls[0].TypeRef)
	}
}
