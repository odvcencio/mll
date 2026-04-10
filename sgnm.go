package mll

import (
	"bytes"
	"fmt"
	"io"
)

// Signature algorithm identifiers.
const (
	SigAlgorithmNone    uint8 = 0
	SigAlgorithmEd25519 uint8 = 1
)

// SgnmSection holds the signature bytes.
type SgnmSection struct {
	KeyIDIdx  uint32 // string table index of the key identifier
	Algorithm uint8
	Signature []byte
}

// Write layout: u32 key_id_idx + u8 algorithm + u32 sig_len + sig_len bytes
func (s SgnmSection) Write(w io.Writer) error {
	if err := WriteUint32LE(w, s.KeyIDIdx); err != nil {
		return err
	}
	if _, err := w.Write([]byte{s.Algorithm}); err != nil {
		return err
	}
	if err := WriteUint32LE(w, uint32(len(s.Signature))); err != nil {
		return err
	}
	if len(s.Signature) > 0 {
		if _, err := w.Write(s.Signature); err != nil {
			return err
		}
	}
	return nil
}

// ReadSgnmSection decodes a SGNM section body.
func ReadSgnmSection(data []byte) (SgnmSection, error) {
	r := bytes.NewReader(data)
	var s SgnmSection
	kBuf, err := readBytes(r, 4)
	if err != nil {
		return SgnmSection{}, fmt.Errorf("mll: SGNM key_id: %w", err)
	}
	s.KeyIDIdx, _ = ReadUint32LE(kBuf)
	aBuf, err := readBytes(r, 1)
	if err != nil {
		return SgnmSection{}, err
	}
	s.Algorithm = aBuf[0]
	lBuf, err := readBytes(r, 4)
	if err != nil {
		return SgnmSection{}, err
	}
	sigLen, _ := ReadUint32LE(lBuf)
	if sigLen > 0 {
		s.Signature, err = readBytes(r, int(sigLen))
		if err != nil {
			return SgnmSection{}, err
		}
	}
	return s, nil
}
