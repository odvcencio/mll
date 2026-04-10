package mll

import (
	"encoding/binary"
	"fmt"
)

// Ref is a typed pointer to a named entity in another section.
// Fixed-width 8 bytes: 4-byte section tag + 4-byte intra-section index.
type Ref struct {
	Tag   [4]byte // section tag of the target section
	Index uint32  // intra-section index of the target entity
}

// Encode returns the 8-byte binary representation of a ref.
func (r Ref) Encode() []byte {
	out := make([]byte, 8)
	copy(out[:4], r.Tag[:])
	binary.LittleEndian.PutUint32(out[4:], r.Index)
	return out
}

// DecodeRef reads a ref from 8 bytes.
func DecodeRef(b []byte) (Ref, error) {
	if len(b) < 8 {
		return Ref{}, fmt.Errorf("mll: ref needs 8 bytes, got %d", len(b))
	}
	var r Ref
	copy(r.Tag[:], b[:4])
	r.Index = binary.LittleEndian.Uint32(b[4:])
	return r, nil
}
