package mll

import (
	"bytes"
	"testing"
)

func TestEntrSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewEntrBuilder()
	b.Add(EntryPoint{NameIdx: strg.Intern("forward"), Kind: EntryKindPipeline})
	b.Add(EntryPoint{
		NameIdx: strg.Intern("embed"),
		Kind:    EntryKindFunction,
		Inputs:  []ValueBinding{{NameIdx: strg.Intern("tokens"), TypeRef: Ref{Tag: TagTYPE, Index: 0}}},
		Outputs: []ValueBinding{{NameIdx: strg.Intern("out"), TypeRef: Ref{Tag: TagTYPE, Index: 1}}},
	})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadEntrSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Entries) != 2 {
		t.Fatalf("count: got %d", len(decoded.Entries))
	}
	if len(decoded.Entries[1].Inputs) != 1 || len(decoded.Entries[1].Outputs) != 1 {
		t.Fatalf("bindings: %+v", decoded.Entries[1])
	}
}
