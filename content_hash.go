package mll

import (
	"bytes"
)

// SealedContentHash computes the BLAKE3-256 content hash for a sealed or
// weights-only artifact per spec §Canonicalization / Sealed file content hash.
//
// The pre-image byte sequence is fixed-width and endianness-unambiguous:
//   - version: 2 bytes, [major, minor] (NOT little-endian u16 — explicit byte order)
//   - profile: 1 byte
//   - file flags with HAS_SIGNATURE forced to 0: 1 byte
//   - for each directory entry in canonical order:
//     tag: 4 bytes, raw
//     flags: 2 bytes, little-endian u16
//     schema_version: 2 bytes, little-endian u16
//     digest: 32 bytes, raw
//
// The pre-image deliberately uses [major, minor] byte order for the version,
// which differs from the on-disk little-endian u16 form (which stores bytes
// as [minor, major]). This is intentional and documented: the content hash
// operates on a semantic byte sequence, not a byte-by-byte copy of the file.
//
// MinReaderMinor, file offsets, section sizes, total file size, padding bytes,
// reserved header bytes, and signature bytes are intentionally excluded.
// MinReaderMinor is a loader policy, not artifact content; the others are
// layout decisions that two conformant writers may make differently without
// changing what the artifact represents.
func SealedContentHash(version Version, profile Profile, fileFlags uint8, entries []DirectoryEntry) Digest {
	canonical := CanonicalSectionOrder(entries, profile)
	var buf bytes.Buffer
	// Version (2 bytes)
	buf.WriteByte(version.Major)
	buf.WriteByte(version.Minor)
	// Profile (1 byte)
	buf.WriteByte(byte(profile))
	// File flags with HAS_SIGNATURE forced to 0
	buf.WriteByte(fileFlags &^ FileFlagHasSignature)
	// Directory entries in canonical order
	for _, e := range canonical {
		buf.Write(e.Tag[:])
		WriteUint16LE(&buf, e.Flags)
		WriteUint16LE(&buf, e.SchemaVersion)
		buf.Write(e.Digest[:])
	}
	return HashBytes(buf.Bytes())
}
