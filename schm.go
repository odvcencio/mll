package mll

import (
	"fmt"
	"io"
)

// SchmSection is the SCHM section. The v1.0 core accepts an empty schema body.
type SchmSection struct{}

// Write encodes an empty SCHM section: u32 count = 0.
func (s SchmSection) Write(w io.Writer) error {
	return WriteUint32LE(w, 0)
}

// ReadSchmSection decodes a SCHM body. The v1.0 core accepts and ignores entries.
func ReadSchmSection(data []byte) (SchmSection, error) {
	if len(data) < 4 {
		return SchmSection{}, fmt.Errorf("mll: SCHM too small")
	}
	// Count is read but ignored by the v1.0 core.
	_, _ = ReadUint32LE(data[:4])
	return SchmSection{}, nil
}
