package mll

import (
	"bytes"
	"testing"
)

func TestMempSectionRoundTrip(t *testing.T) {
	b := NewMempBuilder()
	b.Add(MempEntry{ParamRef: Ref{Tag: TagPARM, Index: 0}, Residency: ResidencyDeviceResident, AccessCount: 1000})
	var buf bytes.Buffer
	b.Write(&buf)
	decoded, err := ReadMempSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Entries) != 1 || decoded.Entries[0].AccessCount != 1000 {
		t.Fatalf("got %+v", decoded)
	}
}
