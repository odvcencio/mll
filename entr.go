package mll

import (
	"bytes"
	"fmt"
	"io"
)

// Entry point kinds.
const (
	EntryKindFunction uint8 = 0
	EntryKindPipeline uint8 = 1
	EntryKindKernel   uint8 = 2
)

// ValueBinding is one input or output slot of an entry point.
type ValueBinding struct {
	NameIdx uint32
	TypeRef Ref
}

// EntryPoint is one entry in the ENTR section.
type EntryPoint struct {
	NameIdx uint32
	Kind    uint8
	Inputs  []ValueBinding
	Outputs []ValueBinding
}

type EntrBuilder struct {
	entries []EntryPoint
}

func NewEntrBuilder() *EntrBuilder { return &EntrBuilder{} }

func (b *EntrBuilder) Add(e EntryPoint) { b.entries = append(b.entries, e) }

// Write encodes the ENTR section body.
// Layout: u32 count + repeat{ u32 name, u8 kind, u32 in_count, ValueBinding[in], u32 out_count, ValueBinding[out] }
// ValueBinding: u32 name_idx + Ref(8) type_ref
func (b *EntrBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.entries))); err != nil {
		return err
	}
	for _, e := range b.entries {
		if err := WriteUint32LE(w, e.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write([]byte{e.Kind}); err != nil {
			return err
		}
		if err := writeValueBindings(w, e.Inputs); err != nil {
			return err
		}
		if err := writeValueBindings(w, e.Outputs); err != nil {
			return err
		}
	}
	return nil
}

func writeValueBindings(w io.Writer, vs []ValueBinding) error {
	if err := WriteUint32LE(w, uint32(len(vs))); err != nil {
		return err
	}
	for _, v := range vs {
		if err := WriteUint32LE(w, v.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write(v.TypeRef.Encode()); err != nil {
			return err
		}
	}
	return nil
}

// EntrSection is the decoded form.
type EntrSection struct {
	Entries []EntryPoint
}

// ReadEntrSection decodes an ENTR section body.
func ReadEntrSection(data []byte) (EntrSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return EntrSection{}, fmt.Errorf("mll: ENTR count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := EntrSection{Entries: make([]EntryPoint, count)}
	for i := uint32(0); i < count; i++ {
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return EntrSection{}, err
		}
		s.Entries[i].NameIdx, _ = ReadUint32LE(nBuf)
		kBuf, err := readBytes(r, 1)
		if err != nil {
			return EntrSection{}, err
		}
		s.Entries[i].Kind = kBuf[0]
		inputs, err := readValueBindings(r)
		if err != nil {
			return EntrSection{}, err
		}
		s.Entries[i].Inputs = inputs
		outputs, err := readValueBindings(r)
		if err != nil {
			return EntrSection{}, err
		}
		s.Entries[i].Outputs = outputs
	}
	return s, nil
}

func readValueBindings(r *bytes.Reader) ([]ValueBinding, error) {
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return nil, err
	}
	count, _ := ReadUint32LE(cBuf)
	out := make([]ValueBinding, count)
	for i := uint32(0); i < count; i++ {
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return nil, err
		}
		out[i].NameIdx, _ = ReadUint32LE(nBuf)
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return nil, err
		}
		ref, err := DecodeRef(refBuf)
		if err != nil {
			return nil, err
		}
		out[i].TypeRef = ref
	}
	return out, nil
}
