package mll

import (
	"bytes"
	"fmt"
	"io"
)

// DimBound kinds for a dimension declaration.
const (
	DimBoundDynamic uint8 = 0 // bound later, at entry or load time
	DimBoundStatic  uint8 = 1 // value is fixed
)

// DimDecl is one entry in the DIMS section.
type DimDecl struct {
	NameIdx uint32 // string table index of the dim name
	Bound   uint8  // DimBoundDynamic | DimBoundStatic
	Value   int64  // meaningful when Bound == DimBoundStatic
}

// DimsBuilder accumulates dim declarations.
type DimsBuilder struct {
	decls []DimDecl
}

// NewDimsBuilder returns an empty builder.
func NewDimsBuilder() *DimsBuilder {
	return &DimsBuilder{}
}

// Add appends a dim declaration.
func (b *DimsBuilder) Add(d DimDecl) {
	b.decls = append(b.decls, d)
}

// Decls returns the current slice (read-only).
func (b *DimsBuilder) Decls() []DimDecl { return b.decls }

// Write encodes the DIMS section body.
// Layout: u32 count + repeat{ u32 name_idx, u8 bound, i64 value (always, ignored for dynamic) }
func (b *DimsBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.decls))); err != nil {
		return err
	}
	for _, d := range b.decls {
		if err := WriteUint32LE(w, d.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write([]byte{d.Bound}); err != nil {
			return err
		}
		if err := WriteUint64LE(w, uint64(d.Value)); err != nil {
			return err
		}
	}
	return nil
}

// DimsSection is the decoded form of the DIMS section.
type DimsSection struct {
	Decls []DimDecl
}

// ReadDimsSection decodes a DIMS section body.
func ReadDimsSection(data []byte) (DimsSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return DimsSection{}, fmt.Errorf("mll: DIMS count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := DimsSection{Decls: make([]DimDecl, count)}
	for i := uint32(0); i < count; i++ {
		nameBuf, err := readBytes(r, 4)
		if err != nil {
			return DimsSection{}, err
		}
		s.Decls[i].NameIdx, _ = ReadUint32LE(nameBuf)
		bBuf, err := readBytes(r, 1)
		if err != nil {
			return DimsSection{}, err
		}
		s.Decls[i].Bound = bBuf[0]
		vBuf, err := readBytes(r, 8)
		if err != nil {
			return DimsSection{}, err
		}
		v, _ := ReadUint64LE(vBuf)
		s.Decls[i].Value = int64(v)
	}
	return s, nil
}
