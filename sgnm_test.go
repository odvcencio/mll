package mll

import (
	"bytes"
	"testing"
)

func TestSgnmSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	orig := SgnmSection{
		KeyIDIdx:  strg.Intern("team-key-1"),
		Algorithm: SigAlgorithmEd25519,
		Signature: []byte{1, 2, 3, 4, 5},
	}
	var buf bytes.Buffer
	if err := orig.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadSgnmSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Algorithm != SigAlgorithmEd25519 || len(decoded.Signature) != 5 {
		t.Fatalf("got %+v", decoded)
	}
}
