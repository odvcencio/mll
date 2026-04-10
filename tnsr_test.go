package mll

import (
	"bytes"
	"testing"
)

func TestTnsrBuilderRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewTnsrBuilder()
	b.Add(TensorEntry{
		NameIdx: strg.Intern("weights"),
		DType:   DTypeF32,
		Shape:   []uint64{2, 3},
		Data:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
	})
	b.Add(TensorEntry{
		NameIdx: strg.Intern("bias"),
		DType:   DTypeF32,
		Shape:   []uint64{3},
		Data:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadTnsrSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Tensors) != 2 {
		t.Fatalf("tensor count: got %d", len(decoded.Tensors))
	}
	if decoded.Tensors[0].DType != DTypeF32 {
		t.Errorf("dtype: got %d", decoded.Tensors[0].DType)
	}
	// Verify data
	if !bytes.Equal(decoded.Tensors[0].Data[:24], b.entries[0].Data[:24]) {
		t.Errorf("weights data mismatch")
	}
}

func TestTnsrAlignmentWithinSection(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewTnsrBuilder()
	b.Add(TensorEntry{
		NameIdx: strg.Intern("a"),
		DType:   DTypeF32,
		Shape:   []uint64{3},
		Data:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	})
	b.Add(TensorEntry{
		NameIdx: strg.Intern("b"),
		DType:   DTypeF32,
		Shape:   []uint64{2},
		Data:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
	})
	var buf bytes.Buffer
	b.Write(&buf)
	// The second tensor's body offset must be 64-aligned within the section body.
	decoded, _ := ReadTnsrSection(buf.Bytes())
	if decoded.Tensors[1].BodyOffset%64 != 0 {
		t.Errorf("second tensor body offset not 64-aligned: %d", decoded.Tensors[1].BodyOffset)
	}
}
