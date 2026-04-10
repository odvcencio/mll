package mll

import (
	"fmt"
	"io"
)

// TypeKind identifies the shape of a TYPE section entry.
const (
	TypeKindTensor        uint8 = 1
	TypeKindKVCache       uint8 = 2
	TypeKindCandidatePack uint8 = 3
)

// TypeDecl is one entry in the TYPE section.
type TypeDecl struct {
	NameIdx uint32
	Kind    uint8
	// Tensor fields (valid when Kind == TypeKindTensor)
	DType DType
	Shape []Dimension
	// KV cache fields (valid when Kind == TypeKindKVCache)
	Layers  uint32
	Heads   uint32
	HeadDim uint32
	// Candidate pack fields (valid when Kind == TypeKindCandidatePack)
	Rank uint32
}

// TypeBuilder accumulates type declarations.
type TypeBuilder struct {
	decls []TypeDecl
}

// NewTypeBuilder returns an empty builder.
func NewTypeBuilder() *TypeBuilder {
	return &TypeBuilder{}
}

// AddTensorType appends a tensor type and returns its index.
func (b *TypeBuilder) AddTensorType(nameIdx uint32, dtype DType, shape []Dimension) uint32 {
	idx := uint32(len(b.decls))
	b.decls = append(b.decls, TypeDecl{
		NameIdx: nameIdx,
		Kind:    TypeKindTensor,
		DType:   dtype,
		Shape:   shape,
	})
	return idx
}

// AddKVCacheType appends a kv-cache type and returns its index.
func (b *TypeBuilder) AddKVCacheType(nameIdx uint32, layers, heads, headDim uint32) uint32 {
	idx := uint32(len(b.decls))
	b.decls = append(b.decls, TypeDecl{
		NameIdx: nameIdx,
		Kind:    TypeKindKVCache,
		Layers:  layers,
		Heads:   heads,
		HeadDim: headDim,
	})
	return idx
}

// AddCandidatePackType appends a candidate-pack type and returns its index.
func (b *TypeBuilder) AddCandidatePackType(nameIdx uint32, rank uint32) uint32 {
	idx := uint32(len(b.decls))
	b.decls = append(b.decls, TypeDecl{
		NameIdx: nameIdx,
		Kind:    TypeKindCandidatePack,
		Rank:    rank,
	})
	return idx
}

// Decls returns the current slice (read-only).
func (b *TypeBuilder) Decls() []TypeDecl { return b.decls }

// Write encodes the TYPE section body.
// Layout: u32 count + repeat{ u32 name_idx, u8 kind, kind-specific payload }
//
//	tensor:         u8 dtype + Shape (u32 rank + Dimension[rank] via WriteShape)
//	kv_cache:       u32 layers + u32 heads + u32 head_dim
//	candidate_pack: u32 rank
func (b *TypeBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.decls))); err != nil {
		return err
	}
	for _, d := range b.decls {
		if err := WriteUint32LE(w, d.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write([]byte{d.Kind}); err != nil {
			return err
		}
		switch d.Kind {
		case TypeKindTensor:
			if _, err := w.Write([]byte{byte(d.DType)}); err != nil {
				return err
			}
			if err := WriteShape(w, d.Shape); err != nil {
				return err
			}
		case TypeKindKVCache:
			if err := WriteUint32LE(w, d.Layers); err != nil {
				return err
			}
			if err := WriteUint32LE(w, d.Heads); err != nil {
				return err
			}
			if err := WriteUint32LE(w, d.HeadDim); err != nil {
				return err
			}
		case TypeKindCandidatePack:
			if err := WriteUint32LE(w, d.Rank); err != nil {
				return err
			}
		default:
			return fmt.Errorf("mll: TYPE invalid kind %d", d.Kind)
		}
	}
	return nil
}

// TypeSection is the decoded form of the TYPE section.
type TypeSection struct {
	Decls []TypeDecl
}

// ReadTypeSection decodes a TYPE section body.
func ReadTypeSection(data []byte) (TypeSection, error) {
	if len(data) < 4 {
		return TypeSection{}, fmt.Errorf("mll: TYPE too small")
	}
	count, _ := ReadUint32LE(data[:4])
	cursor := 4
	s := TypeSection{Decls: make([]TypeDecl, count)}
	for i := uint32(0); i < count; i++ {
		if cursor+5 > len(data) {
			return TypeSection{}, fmt.Errorf("mll: TYPE entry %d header truncated", i)
		}
		nameIdx, _ := ReadUint32LE(data[cursor:])
		cursor += 4
		kind := data[cursor]
		cursor++
		s.Decls[i].NameIdx = nameIdx
		s.Decls[i].Kind = kind
		switch kind {
		case TypeKindTensor:
			if cursor >= len(data) {
				return TypeSection{}, fmt.Errorf("mll: TYPE tensor dtype truncated")
			}
			s.Decls[i].DType = DType(data[cursor])
			cursor++
			shape, n, err := ReadShape(data[cursor:])
			if err != nil {
				return TypeSection{}, fmt.Errorf("mll: TYPE tensor shape: %w", err)
			}
			s.Decls[i].Shape = shape
			cursor += n
		case TypeKindKVCache:
			if cursor+12 > len(data) {
				return TypeSection{}, fmt.Errorf("mll: TYPE kv_cache truncated")
			}
			s.Decls[i].Layers, _ = ReadUint32LE(data[cursor:])
			cursor += 4
			s.Decls[i].Heads, _ = ReadUint32LE(data[cursor:])
			cursor += 4
			s.Decls[i].HeadDim, _ = ReadUint32LE(data[cursor:])
			cursor += 4
		case TypeKindCandidatePack:
			if cursor+4 > len(data) {
				return TypeSection{}, fmt.Errorf("mll: TYPE candidate_pack truncated")
			}
			s.Decls[i].Rank, _ = ReadUint32LE(data[cursor:])
			cursor += 4
		default:
			return TypeSection{}, fmt.Errorf("mll: TYPE invalid kind %d", kind)
		}
	}
	return s, nil
}
