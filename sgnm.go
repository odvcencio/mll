package mll

import (
	"bytes"
	"crypto/ed25519"
	"errors"
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

// NewEd25519SgnmSection signs digest with privateKey and returns an SGNM body model.
func NewEd25519SgnmSection(keyIDIdx uint32, privateKey ed25519.PrivateKey, digest Digest) (SgnmSection, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return SgnmSection{}, fmt.Errorf("mll: ed25519 private key has length %d, want %d", len(privateKey), ed25519.PrivateKeySize)
	}
	return SgnmSection{
		KeyIDIdx:  keyIDIdx,
		Algorithm: SigAlgorithmEd25519,
		Signature: ed25519.Sign(privateKey, digest[:]),
	}, nil
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

// VerifySignature verifies the file's SGNM signature over the sealed content hash.
func (r *Reader) VerifySignature(publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("mll: ed25519 public key has length %d, want %d", len(publicKey), ed25519.PublicKeySize)
	}
	if r.header.Profile != ProfileSealed && r.header.Profile != ProfileWeightsOnly {
		return errors.New("mll: signatures are only valid for sealed and weights-only profiles")
	}
	if r.header.Flags&FileFlagHasSignature == 0 {
		return errors.New("mll: file does not set HAS_SIGNATURE")
	}
	body, ok := r.Section(TagSGNM)
	if !ok {
		return errors.New("mll: file sets HAS_SIGNATURE but has no SGNM section")
	}
	sgnm, err := ReadSgnmSection(body)
	if err != nil {
		return err
	}
	if sgnm.Algorithm != SigAlgorithmEd25519 {
		return fmt.Errorf("mll: unsupported signature algorithm %d", sgnm.Algorithm)
	}
	if len(sgnm.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("mll: ed25519 signature has length %d, want %d", len(sgnm.Signature), ed25519.SignatureSize)
	}
	hash, err := r.ContentHash()
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, hash[:], sgnm.Signature) {
		return errors.New("mll: signature verification failed")
	}
	return nil
}
