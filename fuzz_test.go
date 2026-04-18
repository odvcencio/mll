package mll

import (
	"bytes"
	"testing"
)

func FuzzReadBytes(f *testing.F) {
	a := NewSealedArtifact("fuzz")
	if err := a.AddTensor("w", DTypeF32, []uint64{1}, make([]byte, 4)); err != nil {
		f.Fatal(err)
	}
	data, _, err := a.Marshal()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte{0, 1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		r, err := ReadBytes(data)
		if err != nil {
			return
		}
		_ = r.Validate()
		for _, e := range r.DirectoryEntries() {
			_, _ = r.Section(e.Tag)
		}
	})
}

func FuzzVarintRoundTrip(f *testing.F) {
	for _, seed := range []int64{0, 1, -1, 63, -64, 1 << 20, -(1 << 20)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, v int64) {
		var buf bytes.Buffer
		if err := WriteVarint(&buf, v); err != nil {
			t.Fatal(err)
		}
		got, n, err := ReadVarint(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		if got != v || n != buf.Len() {
			t.Fatalf("got (%d,%d), want (%d,%d)", got, n, v, buf.Len())
		}
	})
}

func FuzzDimensionRoundTrip(f *testing.F) {
	for _, seed := range []int64{0, 1, -1, 42, -1000} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, v int64) {
		var buf bytes.Buffer
		if err := WriteDimension(&buf, DimLiteral(v)); err != nil {
			t.Fatal(err)
		}
		got, n, err := ReadDimension(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		if n != buf.Len() || got.Kind != DimKindLiteral || got.Value != v {
			t.Fatalf("round-trip got %+v consumed=%d want literal %d consumed=%d", got, n, v, buf.Len())
		}
	})
}

func FuzzSectionReaders(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0})
	f.Add([]byte{1, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024 {
			return
		}
		if len(data) >= 4 {
			count, _ := ReadUint32LE(data[:4])
			if count > 32 {
				return
			}
		}
		_, _ = ReadHeadSection(data)
		_, _ = ReadStringTable(data)
		_, _ = ReadEnumSection(data)
		_, _ = ReadDimsSection(data)
		_, _ = ReadTypeSection(data)
		_, _ = ReadParmSection(data)
		_, _ = ReadEntrSection(data)
		_, _ = ReadBuffSection(data)
		_, _ = ReadKrnlSection(data)
		_, _ = ReadPlanSection(data)
		_, _ = ReadMempSection(data)
		_, _ = ReadTnsrSection(data)
		_, _ = ReadOptmSection(data)
		_, _ = ReadSgnmSection(data)
	})
}
