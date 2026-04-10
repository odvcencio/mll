package mll

import (
	"bytes"
	"fmt"
	"io"
)

// ParmDecl is one entry in the PARM section.
type ParmDecl struct {
	NameIdx    uint32
	TypeRef    Ref    // reference into TYPE section
	BindingIdx uint32 // string table index; 0 if no binding
	Trainable  bool
}

// ParmBuilder accumulates parameter declarations.
type ParmBuilder struct {
	decls []ParmDecl
}

// NewParmBuilder returns an empty builder.
func NewParmBuilder() *ParmBuilder {
	return &ParmBuilder{}
}

// Add appends a parameter declaration.
func (b *ParmBuilder) Add(p ParmDecl) {
	b.decls = append(b.decls, p)
}

// Write encodes the PARM section body.
// Layout: u32 count + repeat{ u32 name_idx, Ref(8) type_ref, u32 binding_idx, u8 trainable }
func (b *ParmBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.decls))); err != nil {
		return err
	}
	for _, p := range b.decls {
		if err := WriteUint32LE(w, p.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write(p.TypeRef.Encode()); err != nil {
			return err
		}
		if err := WriteUint32LE(w, p.BindingIdx); err != nil {
			return err
		}
		var tb byte
		if p.Trainable {
			tb = 1
		}
		if _, err := w.Write([]byte{tb}); err != nil {
			return err
		}
	}
	return nil
}

// ParmSection is the decoded form.
type ParmSection struct {
	Decls []ParmDecl
}

// ReadParmSection decodes a PARM section body.
func ReadParmSection(data []byte) (ParmSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return ParmSection{}, fmt.Errorf("mll: PARM count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := ParmSection{Decls: make([]ParmDecl, count)}
	for i := uint32(0); i < count; i++ {
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return ParmSection{}, err
		}
		s.Decls[i].NameIdx, _ = ReadUint32LE(nBuf)
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return ParmSection{}, err
		}
		ref, err := DecodeRef(refBuf)
		if err != nil {
			return ParmSection{}, err
		}
		s.Decls[i].TypeRef = ref
		bBuf, err := readBytes(r, 4)
		if err != nil {
			return ParmSection{}, err
		}
		s.Decls[i].BindingIdx, _ = ReadUint32LE(bBuf)
		tBuf, err := readBytes(r, 1)
		if err != nil {
			return ParmSection{}, err
		}
		s.Decls[i].Trainable = tBuf[0] != 0
	}
	return s, nil
}
