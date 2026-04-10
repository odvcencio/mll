package mll

import (
	"fmt"
	"io"
)

// SchmSection is the SCHM section. In Plan 1 this is always empty —
// full schema support lands in Plan 2.
type SchmSection struct{}

// Write encodes an empty SCHM section: u32 count = 0.
func (s SchmSection) Write(w io.Writer) error {
	return WriteUint32LE(w, 0)
}

// ReadSchmSection decodes a SCHM body. Plan 1 accepts and ignores any entries.
func ReadSchmSection(data []byte) (SchmSection, error) {
	if len(data) < 4 {
		return SchmSection{}, fmt.Errorf("mll: SCHM too small")
	}
	// Count is read but ignored in Plan 1.
	_, _ = ReadUint32LE(data[:4])
	return SchmSection{}, nil
}
