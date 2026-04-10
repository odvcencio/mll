package mll

import "testing"

func TestCustomChunkTagCheck(t *testing.T) {
	c := CustomChunk{Tag: [4]byte{'X', 'M', 'C', 'D'}, Body: []byte{1, 2, 3}}
	if !IsCustomTag(c.Tag) {
		t.Fatal("X-prefixed tags should be custom")
	}
}
