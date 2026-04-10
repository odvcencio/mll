package mll

import (
	"bytes"
	"fmt"
	"io"
)

// Storage classes for buffer declarations.
const (
	StorageClassActivation uint8 = 0
	StorageClassWorkspace  uint8 = 1
	StorageClassIO         uint8 = 2
)

// BuffDecl is one entry in the BUFF section.
type BuffDecl struct {
	NameIdx      uint32
	TypeRef      Ref
	StorageClass uint8
}

type BuffBuilder struct {
	decls []BuffDecl
}

func NewBuffBuilder() *BuffBuilder { return &BuffBuilder{} }

func (b *BuffBuilder) Add(d BuffDecl) { b.decls = append(b.decls, d) }

// Write encodes the BUFF section body.
// Layout: u32 count + repeat{ u32 name, Ref(8), u8 storage_class }
func (b *BuffBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.decls))); err != nil {
		return err
	}
	for _, d := range b.decls {
		if err := WriteUint32LE(w, d.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write(d.TypeRef.Encode()); err != nil {
			return err
		}
		if _, err := w.Write([]byte{d.StorageClass}); err != nil {
			return err
		}
	}
	return nil
}

type BuffSection struct {
	Decls []BuffDecl
}

func ReadBuffSection(data []byte) (BuffSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return BuffSection{}, fmt.Errorf("mll: BUFF count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := BuffSection{Decls: make([]BuffDecl, count)}
	for i := uint32(0); i < count; i++ {
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return BuffSection{}, err
		}
		s.Decls[i].NameIdx, _ = ReadUint32LE(nBuf)
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return BuffSection{}, err
		}
		ref, err := DecodeRef(refBuf)
		if err != nil {
			return BuffSection{}, err
		}
		s.Decls[i].TypeRef = ref
		scBuf, err := readBytes(r, 1)
		if err != nil {
			return BuffSection{}, err
		}
		s.Decls[i].StorageClass = scBuf[0]
	}
	return s, nil
}
