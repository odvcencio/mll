package mll

import (
	"bytes"
	"fmt"
	"io"
)

// KernelDecl is a minimal kernel declaration. The full DSL (tile dims, variants,
// op bodies) lives in Plan 2+; Plan 1 stores the kernel name and an opaque
// body byte blob as a placeholder so callers can round-trip a KRNL section.
type KernelDecl struct {
	NameIdx uint32
	Body    []byte
}

type KrnlBuilder struct {
	decls []KernelDecl
}

func NewKrnlBuilder() *KrnlBuilder { return &KrnlBuilder{} }

func (b *KrnlBuilder) Add(k KernelDecl) { b.decls = append(b.decls, k) }

// Write layout: u32 count + repeat{ u32 name, u32 body_len, body_len bytes }
func (b *KrnlBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.decls))); err != nil {
		return err
	}
	for _, k := range b.decls {
		if err := WriteUint32LE(w, k.NameIdx); err != nil {
			return err
		}
		if err := WriteUint32LE(w, uint32(len(k.Body))); err != nil {
			return err
		}
		if len(k.Body) > 0 {
			if _, err := w.Write(k.Body); err != nil {
				return err
			}
		}
	}
	return nil
}

type KrnlSection struct {
	Decls []KernelDecl
}

func ReadKrnlSection(data []byte) (KrnlSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return KrnlSection{}, fmt.Errorf("mll: KRNL count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := KrnlSection{Decls: make([]KernelDecl, count)}
	for i := uint32(0); i < count; i++ {
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return KrnlSection{}, err
		}
		s.Decls[i].NameIdx, _ = ReadUint32LE(nBuf)
		lBuf, err := readBytes(r, 4)
		if err != nil {
			return KrnlSection{}, err
		}
		length, _ := ReadUint32LE(lBuf)
		if length > 0 {
			body, err := readBytes(r, int(length))
			if err != nil {
				return KrnlSection{}, err
			}
			s.Decls[i].Body = body
		}
	}
	return s, nil
}
