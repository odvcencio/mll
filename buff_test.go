package mll

import (
	"bytes"
	"testing"
)

func TestBuffSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewBuffBuilder()
	b.Add(BuffDecl{NameIdx: strg.Intern("workspace"), TypeRef: Ref{Tag: TagTYPE, Index: 0}, StorageClass: StorageClassWorkspace})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadBuffSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Decls) != 1 || decoded.Decls[0].StorageClass != StorageClassWorkspace {
		t.Fatalf("got %+v", decoded)
	}
}
