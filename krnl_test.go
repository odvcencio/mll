package mll

import (
	"bytes"
	"testing"
)

func TestKrnlSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewKrnlBuilder()
	b.Add(KernelDecl{NameIdx: strg.Intern("matmul_f32"), Body: []byte{0x01, 0x02, 0x03}})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadKrnlSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 1 || len(decoded.Decls[0].Body) != 3 {
		t.Fatalf("got %+v", decoded)
	}
}

func TestKrnlSectionEmpty(t *testing.T) {
	b := NewKrnlBuilder()
	var buf bytes.Buffer
	b.Write(&buf)
	decoded, err := ReadKrnlSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 0 {
		t.Fatalf("got %+v", decoded)
	}
}
