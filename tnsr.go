package mll

import (
	"bytes"
	"fmt"
	"io"
)

// TensorEntry describes one tensor in the TNSR section.
type TensorEntry struct {
	NameIdx    uint32 // string table index for tensor name
	DType      DType
	Shape      []uint64
	BodyOffset uint64 // offset within the TNSR section body (computed by builder)
	BodySize   uint64 // byte length of tensor data (computed by builder)
	Data       []byte // raw bytes (used during build; read-only during load)
}

// TnsrSection is the TNSR section body.
type TnsrSection struct {
	Tensors []TensorEntry
}

// TnsrBuilder accumulates tensor entries and writes a properly aligned section.
type TnsrBuilder struct {
	entries []TensorEntry
}

// NewTnsrBuilder creates an empty builder.
func NewTnsrBuilder() *TnsrBuilder {
	return &TnsrBuilder{}
}

// Add appends a tensor entry.
func (b *TnsrBuilder) Add(e TensorEntry) {
	b.entries = append(b.entries, e)
}

// Write encodes the TNSR section body to w. Tensor bodies are 64-byte aligned
// within the section. The caller is responsible for setting the ALIGNED flag
// on the directory entry so the section body itself is page-aligned in the file.
//
// Section body layout:
//
//	u32 tensor_count
//	for each tensor:
//	    u32 name_idx
//	    u8  dtype
//	    u32 rank
//	    u64[rank] shape
//	    u64 body_offset  (filled in during write)
//	    u64 body_size
//	    u8  flags
//	    u8[3] pad
//	[alignment pad to 64-byte boundary]
//	[tensor 1 raw bytes]
//	[alignment pad]
//	[tensor 2 raw bytes]
//	...
func (b *TnsrBuilder) Write(w io.Writer) error {
	// First pass: compute header size and tensor offsets.
	headerSize := uint64(4)
	for _, e := range b.entries {
		// per-entry header: 4 (name) + 1 (dtype) + 4 (rank) + 8*rank (shape) + 8 (offset) + 8 (size) + 1 (flags) + 3 (pad)
		headerSize += 4 + 1 + 4 + 8*uint64(len(e.Shape)) + 8 + 8 + 1 + 3
	}
	// Pad header to 64-byte boundary
	const align = 64
	headerEnd := headerSize
	if rem := headerEnd % align; rem != 0 {
		headerEnd += align - rem
	}
	// Assign body offsets
	cursor := headerEnd
	offsets := make([]uint64, len(b.entries))
	sizes := make([]uint64, len(b.entries))
	for i, e := range b.entries {
		offsets[i] = cursor
		sizes[i] = uint64(len(e.Data))
		cursor += sizes[i]
		// Pad each tensor to 64-byte boundary before the next
		if rem := cursor % align; rem != 0 {
			cursor += align - rem
		}
	}
	// Now write the header.
	var buf bytes.Buffer
	WriteUint32LE(&buf, uint32(len(b.entries)))
	for i, e := range b.entries {
		WriteUint32LE(&buf, e.NameIdx)
		buf.WriteByte(byte(e.DType))
		WriteUint32LE(&buf, uint32(len(e.Shape)))
		for _, d := range e.Shape {
			WriteUint64LE(&buf, d)
		}
		WriteUint64LE(&buf, offsets[i])
		WriteUint64LE(&buf, sizes[i])
		buf.WriteByte(0)           // flags
		buf.Write([]byte{0, 0, 0}) // pad
	}
	if uint64(buf.Len()) != headerSize {
		return fmt.Errorf("mll: tnsr header size mismatch: got %d, want %d", buf.Len(), headerSize)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}
	// Header alignment pad
	if pad := headerEnd - headerSize; pad > 0 {
		if _, err := w.Write(make([]byte, pad)); err != nil {
			return err
		}
	}
	// Write tensor bodies with alignment pads.
	written := headerEnd
	for i, e := range b.entries {
		if written != offsets[i] {
			return fmt.Errorf("mll: tnsr body offset mismatch for tensor %d", i)
		}
		if _, err := w.Write(e.Data); err != nil {
			return err
		}
		written += sizes[i]
		if rem := written % align; rem != 0 {
			pad := align - rem
			if _, err := w.Write(make([]byte, pad)); err != nil {
				return err
			}
			written += pad
		}
	}
	return nil
}

// ReadTnsrSection parses a TNSR section body.
func ReadTnsrSection(data []byte) (TnsrSection, error) {
	if len(data) < 4 {
		return TnsrSection{}, fmt.Errorf("mll: tnsr body too small")
	}
	count, _ := ReadUint32LE(data[:4])
	cursor := 4
	s := TnsrSection{Tensors: make([]TensorEntry, count)}
	for i := uint32(0); i < count; i++ {
		if cursor+4+1+4 > len(data) {
			return TnsrSection{}, fmt.Errorf("mll: tnsr tensor %d header truncated", i)
		}
		nameIdx, _ := ReadUint32LE(data[cursor:])
		cursor += 4
		dtype := DType(data[cursor])
		cursor += 1
		rank, _ := ReadUint32LE(data[cursor:])
		cursor += 4
		if rank > uint32((len(data)-cursor)/8) {
			return TnsrSection{}, fmt.Errorf("mll: tnsr tensor %d shape truncated", i)
		}
		shape := make([]uint64, rank)
		for r := uint32(0); r < rank; r++ {
			shape[r], _ = ReadUint64LE(data[cursor:])
			cursor += 8
		}
		if cursor+8+8+1+3 > len(data) {
			return TnsrSection{}, fmt.Errorf("mll: tnsr tensor %d metadata truncated", i)
		}
		bodyOffset, _ := ReadUint64LE(data[cursor:])
		cursor += 8
		bodySize, _ := ReadUint64LE(data[cursor:])
		cursor += 8
		cursor += 1 + 3 // flags + pad
		s.Tensors[i] = TensorEntry{
			NameIdx:    nameIdx,
			DType:      dtype,
			Shape:      shape,
			BodyOffset: bodyOffset,
			BodySize:   bodySize,
		}
	}
	// Fill in tensor data views.
	for i := range s.Tensors {
		off := s.Tensors[i].BodyOffset
		size := s.Tensors[i].BodySize
		if off+size > uint64(len(data)) {
			return TnsrSection{}, fmt.Errorf("mll: tnsr tensor %d body out of bounds", i)
		}
		s.Tensors[i].Data = data[off : off+size]
	}
	return s, nil
}
