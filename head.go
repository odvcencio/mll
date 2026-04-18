package mll

import (
	"bytes"
	"fmt"
	"io"
)

// HeadValueKind is the private enum for HEAD metadata values.
// This enum is NOT the same as the global primitive KindXxx enum —
// HEAD metadata is restricted to scalar primitives.
type HeadValueKind uint8

const (
	HeadValueNull   HeadValueKind = 0
	HeadValueBool   HeadValueKind = 1
	HeadValueI64    HeadValueKind = 2
	HeadValueF64    HeadValueKind = 3
	HeadValueString HeadValueKind = 4
)

// HeadMetadataEntry is one typed key-value in HEAD metadata.
type HeadMetadataEntry struct {
	Key       uint32 // string table index
	Kind      HeadValueKind
	Bool      bool
	I64       int64
	F64       float64
	StringIdx uint32
}

// HeadSection is the HEAD section body.
// Field semantics match the v1.0 HEAD section.
type HeadSection struct {
	Name          uint32 // string table index (required)
	Description   uint32 // string table index (0 = absent)
	CreatedUnixMs int64
	Generation    uint64 // checkpoint only; zero for sealed and weights-only
	Backends      []uint16
	Capabilities  []uint32
	Metadata      []HeadMetadataEntry
}

// Write encodes the HEAD section body to w.
func (h HeadSection) Write(w io.Writer) error {
	if err := WriteUint32LE(w, h.Name); err != nil {
		return err
	}
	if err := WriteUint32LE(w, h.Description); err != nil {
		return err
	}
	if err := WriteUint64LE(w, uint64(h.CreatedUnixMs)); err != nil {
		return err
	}
	if err := WriteUint64LE(w, h.Generation); err != nil {
		return err
	}
	if err := WriteUint16LE(w, uint16(len(h.Backends))); err != nil {
		return err
	}
	for _, b := range h.Backends {
		if err := WriteUint16LE(w, b); err != nil {
			return err
		}
	}
	if err := WriteUint16LE(w, uint16(len(h.Capabilities))); err != nil {
		return err
	}
	for _, c := range h.Capabilities {
		if err := WriteUint32LE(w, c); err != nil {
			return err
		}
	}
	if err := WriteUint16LE(w, uint16(len(h.Metadata))); err != nil {
		return err
	}
	for _, entry := range h.Metadata {
		if err := WriteUint32LE(w, entry.Key); err != nil {
			return err
		}
		if _, err := w.Write([]byte{byte(entry.Kind)}); err != nil {
			return err
		}
		switch entry.Kind {
		case HeadValueNull:
			// no payload
		case HeadValueBool:
			var b byte
			if entry.Bool {
				b = 1
			}
			if _, err := w.Write([]byte{b}); err != nil {
				return err
			}
		case HeadValueI64:
			if err := WriteUint64LE(w, uint64(entry.I64)); err != nil {
				return err
			}
		case HeadValueF64:
			if err := WriteUint64LE(w, Float64bits(entry.F64)); err != nil {
				return err
			}
		case HeadValueString:
			if err := WriteUint32LE(w, entry.StringIdx); err != nil {
				return err
			}
		default:
			return fmt.Errorf("mll: HEAD metadata invalid kind %d", entry.Kind)
		}
	}
	return nil
}

// ReadHeadSection decodes a HEAD section body from b.
func ReadHeadSection(b []byte) (HeadSection, error) {
	r := bytes.NewReader(b)
	var h HeadSection
	nameBuf, err := readBytes(r, 4)
	if err != nil {
		return h, err
	}
	h.Name, _ = ReadUint32LE(nameBuf)
	descBuf, err := readBytes(r, 4)
	if err != nil {
		return h, err
	}
	h.Description, _ = ReadUint32LE(descBuf)
	createdBuf, err := readBytes(r, 8)
	if err != nil {
		return h, err
	}
	createdU64, _ := ReadUint64LE(createdBuf)
	h.CreatedUnixMs = int64(createdU64)
	genBuf, err := readBytes(r, 8)
	if err != nil {
		return h, err
	}
	h.Generation, _ = ReadUint64LE(genBuf)
	bCountBuf, err := readBytes(r, 2)
	if err != nil {
		return h, err
	}
	bCount, _ := ReadUint16LE(bCountBuf)
	h.Backends = make([]uint16, bCount)
	for i := range h.Backends {
		buf, err := readBytes(r, 2)
		if err != nil {
			return h, err
		}
		h.Backends[i], _ = ReadUint16LE(buf)
	}
	cCountBuf, err := readBytes(r, 2)
	if err != nil {
		return h, err
	}
	cCount, _ := ReadUint16LE(cCountBuf)
	h.Capabilities = make([]uint32, cCount)
	for i := range h.Capabilities {
		buf, err := readBytes(r, 4)
		if err != nil {
			return h, err
		}
		h.Capabilities[i], _ = ReadUint32LE(buf)
	}
	mCountBuf, err := readBytes(r, 2)
	if err != nil {
		return h, err
	}
	mCount, _ := ReadUint16LE(mCountBuf)
	h.Metadata = make([]HeadMetadataEntry, mCount)
	for i := range h.Metadata {
		keyBuf, err := readBytes(r, 4)
		if err != nil {
			return h, err
		}
		h.Metadata[i].Key, _ = ReadUint32LE(keyBuf)
		kindBuf, err := readBytes(r, 1)
		if err != nil {
			return h, err
		}
		h.Metadata[i].Kind = HeadValueKind(kindBuf[0])
		switch h.Metadata[i].Kind {
		case HeadValueNull:
			// no payload
		case HeadValueBool:
			boolBuf, err := readBytes(r, 1)
			if err != nil {
				return h, err
			}
			h.Metadata[i].Bool = boolBuf[0] != 0
		case HeadValueI64:
			buf, err := readBytes(r, 8)
			if err != nil {
				return h, err
			}
			v, _ := ReadUint64LE(buf)
			h.Metadata[i].I64 = int64(v)
		case HeadValueF64:
			buf, err := readBytes(r, 8)
			if err != nil {
				return h, err
			}
			bits, _ := ReadUint64LE(buf)
			h.Metadata[i].F64 = Float64frombits(bits)
		case HeadValueString:
			buf, err := readBytes(r, 4)
			if err != nil {
				return h, err
			}
			h.Metadata[i].StringIdx, _ = ReadUint32LE(buf)
		default:
			return h, fmt.Errorf("mll: HEAD metadata unknown kind %d", h.Metadata[i].Kind)
		}
	}
	return h, nil
}

// readBytes is a helper that reads exactly n bytes from r.
func readBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

// DigestBody returns the byte sequence that should be hashed for this HEAD
// section under the given profile. For sealed and weights-only profiles,
// created_unix_ms and generation are zeroed so reproducible builds work
// across different wall clocks. For checkpoint profile, the full body is used.
func (h HeadSection) DigestBody(profile Profile) []byte {
	if profile == ProfileCheckpoint {
		var buf bytes.Buffer
		h.Write(&buf)
		return buf.Bytes()
	}
	// Sealed/weights-only: zero out wall-clock fields.
	clone := h
	clone.CreatedUnixMs = 0
	clone.Generation = 0
	var buf bytes.Buffer
	clone.Write(&buf)
	return buf.Bytes()
}
