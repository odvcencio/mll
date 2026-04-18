package mll

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// SealedArtifact is a convenience builder for simple sealed tensor artifacts.
// It keeps the low-level section model available while covering the common
// case where callers need HEAD, STRG, DIMS, TYPE, PARM, ENTR, and TNSR.
type SealedArtifact struct {
	name    string
	dims    []sealedArtifactDim
	tensors []sealedArtifactTensor
}

type sealedArtifactDim struct {
	name  string
	value int64
}

type sealedArtifactTensor struct {
	name  string
	dtype DType
	shape []uint64
	data  []byte
}

// NewSealedArtifact creates a sealed artifact builder with the given name.
func NewSealedArtifact(name string) *SealedArtifact {
	return &SealedArtifact{name: name}
}

// AddDim adds a named static dimension declaration.
func (a *SealedArtifact) AddDim(name string, value int64) {
	a.dims = append(a.dims, sealedArtifactDim{name: name, value: value})
}

// AddTensor adds a tensor parameter and its raw bytes.
func (a *SealedArtifact) AddTensor(name string, dtype DType, shape []uint64, data []byte) error {
	if err := validateArtifactTensorData(dtype, shape, data); err != nil {
		return err
	}
	a.tensors = append(a.tensors, sealedArtifactTensor{
		name:  name,
		dtype: dtype,
		shape: append([]uint64(nil), shape...),
		data:  append([]byte(nil), data...),
	})
	return nil
}

// Marshal writes the artifact and returns the file bytes plus sealed content hash.
func (a *SealedArtifact) Marshal() ([]byte, Digest, error) {
	strg := NewStringTableBuilder()
	strg.Intern(a.name)
	strg.Intern("forward")
	for _, d := range a.dims {
		strg.Intern(d.name)
	}
	for _, t := range a.tensors {
		strg.Intern(t.name)
		strg.Intern(t.name + ".type")
	}
	strg.CanonicalizeLexicographic()

	lookup := func(s string) uint32 {
		idx, ok := strg.Lookup(s)
		if !ok {
			panic("mll: internal string table lookup failed")
		}
		return idx
	}

	dims := NewDimsBuilder()
	for _, d := range a.dims {
		dims.Add(DimDecl{NameIdx: lookup(d.name), Bound: DimBoundStatic, Value: d.value})
	}

	types := NewTypeBuilder()
	for _, t := range a.tensors {
		shape := make([]Dimension, len(t.shape))
		for i, dim := range t.shape {
			if dim > math.MaxInt64 {
				return nil, Digest{}, fmt.Errorf("mll: tensor %q dimension %d exceeds int64", t.name, dim)
			}
			shape[i] = DimLiteral(int64(dim))
		}
		types.AddTensorType(lookup(t.name+".type"), t.dtype, shape)
	}

	parm := NewParmBuilder()
	for i, t := range a.tensors {
		parm.Add(ParmDecl{
			NameIdx:   lookup(t.name),
			TypeRef:   Ref{Tag: TagTYPE, Index: uint32(i)},
			Trainable: false,
		})
	}

	entr := NewEntrBuilder()
	entr.Add(EntryPoint{NameIdx: lookup("forward"), Kind: EntryKindPipeline})

	tnsr := NewTnsrBuilder()
	for _, t := range a.tensors {
		tnsr.Add(TensorEntry{
			NameIdx: lookup(t.name),
			DType:   t.dtype,
			Shape:   append([]uint64(nil), t.shape...),
			Data:    append([]byte(nil), t.data...),
		})
	}

	head := HeadSection{Name: lookup(a.name)}
	headBody, err := encodeSectionBody(head)
	if err != nil {
		return nil, Digest{}, err
	}
	strgBody, err := encodeSectionBody(strg)
	if err != nil {
		return nil, Digest{}, err
	}
	dimsBody, err := encodeSectionBody(dims)
	if err != nil {
		return nil, Digest{}, err
	}
	typeBody, err := encodeSectionBody(types)
	if err != nil {
		return nil, Digest{}, err
	}
	parmBody, err := encodeSectionBody(parm)
	if err != nil {
		return nil, Digest{}, err
	}
	entrBody, err := encodeSectionBody(entr)
	if err != nil {
		return nil, Digest{}, err
	}
	tnsrBody, err := encodeSectionBody(tnsr)
	if err != nil {
		return nil, Digest{}, err
	}

	sections := []SectionInput{
		{Tag: TagHEAD, Body: headBody, DigestBody: head.DigestBody(ProfileSealed), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagSTRG, Body: strgBody, Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagDIMS, Body: dimsBody, Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagTYPE, Body: typeBody, SchemaVersion: 1},
		{Tag: TagPARM, Body: parmBody, Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagENTR, Body: entrBody, Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagTNSR, Body: tnsrBody, Flags: SectionFlagRequired | SectionFlagAligned, SchemaVersion: 1},
	}

	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0)
	for _, section := range sections {
		wr.AddSection(section)
	}
	if err := wr.Finish(); err != nil {
		return nil, Digest{}, err
	}
	return out.Bytes(), wr.ContentHash(), nil
}

func validateArtifactTensorData(dtype DType, shape []uint64, data []byte) error {
	elemSize := dtype.ElementSize()
	if elemSize == 0 {
		return nil
	}
	elements := uint64(1)
	for _, dim := range shape {
		if dim != 0 && elements > math.MaxUint64/dim {
			return fmt.Errorf("mll: tensor shape overflows uint64 element count")
		}
		elements *= dim
	}
	if elements > math.MaxUint64/uint64(elemSize) {
		return fmt.Errorf("mll: tensor byte size overflows uint64")
	}
	want := elements * uint64(elemSize)
	if uint64(len(data)) != want {
		return fmt.Errorf("mll: tensor data has %d bytes, want %d", len(data), want)
	}
	return nil
}

func encodeSectionBody(section interface{ Write(io.Writer) error }) ([]byte, error) {
	var buf bytes.Buffer
	if err := section.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
