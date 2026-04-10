package mll

import (
	"bytes"
	"fmt"
	"io"
)

// Residency classes for MEMP entries.
const (
	ResidencyDeviceResident uint8 = 0
	ResidencyHostPinned     uint8 = 1
	ResidencyHostShared     uint8 = 2
	ResidencyLazyStaged     uint8 = 3
)

// MempEntry is one per-weight residency entry.
type MempEntry struct {
	ParamRef    Ref
	Residency   uint8
	AccessCount uint32
}

type MempBuilder struct {
	entries []MempEntry
}

func NewMempBuilder() *MempBuilder { return &MempBuilder{} }

func (b *MempBuilder) Add(e MempEntry) { b.entries = append(b.entries, e) }

// Write layout: u32 count + repeat{ Ref(8), u8 residency, u32 access_count }
func (b *MempBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.entries))); err != nil {
		return err
	}
	for _, e := range b.entries {
		if _, err := w.Write(e.ParamRef.Encode()); err != nil {
			return err
		}
		if _, err := w.Write([]byte{e.Residency}); err != nil {
			return err
		}
		if err := WriteUint32LE(w, e.AccessCount); err != nil {
			return err
		}
	}
	return nil
}

type MempSection struct {
	Entries []MempEntry
}

func ReadMempSection(data []byte) (MempSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return MempSection{}, fmt.Errorf("mll: MEMP count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := MempSection{Entries: make([]MempEntry, count)}
	for i := uint32(0); i < count; i++ {
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return MempSection{}, err
		}
		ref, _ := DecodeRef(refBuf)
		s.Entries[i].ParamRef = ref
		rBuf, err := readBytes(r, 1)
		if err != nil {
			return MempSection{}, err
		}
		s.Entries[i].Residency = rBuf[0]
		acBuf, err := readBytes(r, 4)
		if err != nil {
			return MempSection{}, err
		}
		s.Entries[i].AccessCount, _ = ReadUint32LE(acBuf)
	}
	return s, nil
}
