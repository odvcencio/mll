package mll

import (
	"bytes"
	"io"
)

// EnumSectionEntry is one enum declaration in the ENUM section.
// All fields are string table indices.
type EnumSectionEntry struct {
	Name   uint32
	Values []uint32
}

// EnumSection is the ENUM section body.
type EnumSection struct {
	Enums []EnumSectionEntry
}

// Write encodes the ENUM section body.
func (e EnumSection) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(e.Enums))); err != nil {
		return err
	}
	for _, entry := range e.Enums {
		if err := WriteUint32LE(w, entry.Name); err != nil {
			return err
		}
		if err := WriteUint32LE(w, uint32(len(entry.Values))); err != nil {
			return err
		}
		for _, v := range entry.Values {
			if err := WriteUint32LE(w, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadEnumSection decodes an ENUM section body.
func ReadEnumSection(b []byte) (EnumSection, error) {
	r := bytes.NewReader(b)
	countBuf, err := readBytes(r, 4)
	if err != nil {
		return EnumSection{}, err
	}
	count, _ := ReadUint32LE(countBuf)
	e := EnumSection{Enums: make([]EnumSectionEntry, count)}
	for i := range e.Enums {
		nameBuf, err := readBytes(r, 4)
		if err != nil {
			return EnumSection{}, err
		}
		e.Enums[i].Name, _ = ReadUint32LE(nameBuf)
		valCountBuf, err := readBytes(r, 4)
		if err != nil {
			return EnumSection{}, err
		}
		valCount, _ := ReadUint32LE(valCountBuf)
		e.Enums[i].Values = make([]uint32, valCount)
		for j := uint32(0); j < valCount; j++ {
			vBuf, err := readBytes(r, 4)
			if err != nil {
				return EnumSection{}, err
			}
			e.Enums[i].Values[j], _ = ReadUint32LE(vBuf)
		}
	}
	return e, nil
}
