package mll

import (
	"bytes"
	"testing"
)

func TestEnumSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	orig := EnumSection{
		Enums: []EnumSectionEntry{
			{
				Name: strg.Intern("BackendKind"),
				Values: []uint32{
					strg.Intern("cuda"),
					strg.Intern("metal"),
				},
			},
			{
				Name: strg.Intern("StepKind"),
				Values: []uint32{
					strg.Intern("matmul"),
					strg.Intern("softmax"),
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := orig.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadEnumSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Enums) != 2 {
		t.Fatalf("enum count: got %d", len(decoded.Enums))
	}
	if len(decoded.Enums[0].Values) != 2 {
		t.Fatalf("first enum value count: got %d", len(decoded.Enums[0].Values))
	}
}
