package mll

import (
	"bytes"
	"testing"
)

func TestSchmSectionEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := (SchmSection{}).Write(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSchmSection(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}
