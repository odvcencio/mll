package mll

import "lukechampine.com/blake3"

// Digest is a BLAKE3-256 hash (32 bytes).
type Digest [32]byte

// HashBytes computes the BLAKE3-256 hash of the given bytes.
func HashBytes(data []byte) Digest {
	return Digest(blake3.Sum256(data))
}

// Hasher is an incremental BLAKE3-256 hasher.
type Hasher struct {
	h *blake3.Hasher
}

// NewHasher returns a new BLAKE3-256 incremental hasher.
func NewHasher() *Hasher {
	return &Hasher{h: blake3.New(32, nil)}
}

// Write adds bytes to the hash state. Always returns nil error.
func (h *Hasher) Write(p []byte) (int, error) {
	return h.h.Write(p)
}

// Sum finalizes the hash and returns the 32-byte digest.
func (h *Hasher) Sum() Digest {
	var out Digest
	h.h.Sum(out[:0])
	return out
}

// Reset clears the hash state for reuse.
func (h *Hasher) Reset() {
	h.h.Reset()
}
