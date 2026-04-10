package mll

import (
	"bytes"
	"testing"
)

func TestStringTableBuilder(t *testing.T) {
	tbl := NewStringTableBuilder()
	idx1 := tbl.Intern("hello")
	idx2 := tbl.Intern("world")
	idx3 := tbl.Intern("hello") // dedupe
	if idx1 == idx2 {
		t.Fatal("different strings should have different indices")
	}
	if idx1 != idx3 {
		t.Fatal("same string should have same index")
	}
}

func TestStringTableEncodeDecode(t *testing.T) {
	tbl := NewStringTableBuilder()
	tbl.Intern("apple")
	tbl.Intern("banana")
	tbl.Intern("cherry")

	var buf bytes.Buffer
	if err := tbl.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}

	loaded, err := ReadStringTable(buf.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if loaded.Size() != tbl.Size() {
		t.Fatalf("size: got %d, want %d", loaded.Size(), tbl.Size())
	}
	// Verify strings round-trip. Canonical ordering is applied on seal,
	// not on Write alone, so we check presence not order at this layer.
	for i, want := range []string{"apple", "banana", "cherry"} {
		_ = i
		if !loaded.Contains(want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestStringTableCanonicalOrder(t *testing.T) {
	tbl := NewStringTableBuilder()
	// insert out of lex order
	tbl.Intern("banana")
	tbl.Intern("apple")
	tbl.Intern("cherry")

	tbl.CanonicalizeLexicographic()

	// After canonicalization, strings in slot order are apple, banana, cherry.
	expected := []string{"apple", "banana", "cherry"}
	for i, want := range expected {
		got := tbl.At(uint32(i))
		if got != want {
			t.Errorf("slot %d: got %q, want %q", i, got, want)
		}
	}
}
