package mll

import (
	"bytes"
	"fmt"
	"io"
)

// SectionInput describes one section the caller wants to include in the file.
// If DigestBody is non-nil, it is used for the section's BLAKE3 digest
// computation instead of Body. This is needed for HEAD under sealed and
// weights-only profiles, where wall-clock fields are zeroed in the digest
// but preserved in the on-disk body.
type SectionInput struct {
	Tag           [4]byte
	Body          []byte
	DigestBody    []byte // optional; falls back to Body if nil
	Flags         uint16
	SchemaVersion uint16
}

// Writer composes an MLL binary file.
type Writer struct {
	out                  io.Writer
	profile              Profile
	version              Version
	fileFlags            uint8
	sections             []SectionInput
	contentHash          Digest
	skipRequirementCheck bool
}

// WriterOption configures Writer behavior.
type WriterOption func(*Writer)

// WithSkipRequirementCheck disables the profile required-section check.
// Intended for test vector generators and unit tests that need to produce
// minimal files without assembling every required section.
func WithSkipRequirementCheck() WriterOption {
	return func(w *Writer) {
		w.skipRequirementCheck = true
	}
}

// NewWriter constructs a Writer that will emit an MLL file of the given profile.
func NewWriter(out io.Writer, profile Profile, version Version, opts ...WriterOption) *Writer {
	w := &Writer{
		out:     out,
		profile: profile,
		version: version,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// SetFileFlag sets a file-level flag bit.
func (w *Writer) SetFileFlag(flag uint8) {
	w.fileFlags |= flag
}

// AddSection appends a section to be written.
func (w *Writer) AddSection(s SectionInput) {
	w.sections = append(w.sections, s)
}

// Finish writes the complete file to the output writer.
func (w *Writer) Finish() error {
	// Validate profile rules: forbidden sections.
	for _, s := range w.sections {
		if IsForbidden(w.profile, s.Tag) {
			return fmt.Errorf("mll: section %v is forbidden in profile %d", s.Tag, w.profile)
		}
	}
	// Validate profile rules: required sections.
	if !w.skipRequirementCheck {
		present := make(map[[4]byte]bool)
		for _, s := range w.sections {
			present[s.Tag] = true
		}
		for tag := range profileRules[w.profile] {
			if IsRequired(w.profile, tag) && !present[tag] {
				return fmt.Errorf("mll: profile %d requires section %v", w.profile, tag)
			}
		}
	}
	// Build directory entries with digests (offsets come later).
	// When a section provides a DigestBody (e.g., HEAD under sealed/weights-only),
	// use that for the hash; otherwise hash the on-disk Body.
	entries := make([]DirectoryEntry, len(w.sections))
	for i, s := range w.sections {
		digestInput := s.Body
		if s.DigestBody != nil {
			digestInput = s.DigestBody
		}
		entries[i] = DirectoryEntry{
			Tag:           s.Tag,
			Size:          uint64(len(s.Body)),
			Digest:        HashBytes(digestInput),
			Flags:         s.Flags,
			SchemaVersion: s.SchemaVersion,
		}
	}
	// Reorder sections to canonical order (sealed and weights-only only).
	ordered := CanonicalSectionOrder(entries, w.profile)
	// Build a parallel slice of bodies in the same order.
	orderedBodies := make([][]byte, len(ordered))
	for i, e := range ordered {
		for _, s := range w.sections {
			if s.Tag == e.Tag {
				orderedBodies[i] = s.Body
				break
			}
		}
	}
	// Compute offsets, honoring ALIGNED flag.
	dirSize := uint64(len(ordered)) * DirectoryEntrySize
	cursor := uint64(HeaderSize) + dirSize
	for i := range ordered {
		if ordered[i].Flags&SectionFlagAligned != 0 {
			const pageSize = 4096
			if rem := cursor % pageSize; rem != 0 {
				cursor += pageSize - rem
			}
		}
		ordered[i].Offset = cursor
		cursor += ordered[i].Size
	}
	totalFileSize := cursor
	// Compute the sealed content hash for sealed/weights-only.
	if w.profile == ProfileSealed || w.profile == ProfileWeightsOnly {
		w.contentHash = SealedContentHash(w.version, w.profile, w.fileFlags, ordered)
	}
	// Write the header.
	hdr := FileHeader{
		Version:       w.version,
		Profile:       w.profile,
		Flags:         w.fileFlags,
		TotalFileSize: totalFileSize,
		SectionCount:  uint32(len(ordered)),
	}
	if err := hdr.Write(w.out); err != nil {
		return err
	}
	// Write the directory.
	if err := WriteDirectory(w.out, ordered); err != nil {
		return err
	}
	// Write section bodies with alignment padding as needed.
	written := uint64(HeaderSize) + dirSize
	for i, e := range ordered {
		if e.Flags&SectionFlagAligned != 0 {
			const pageSize = 4096
			if rem := written % pageSize; rem != 0 {
				pad := pageSize - rem
				if _, err := w.out.Write(make([]byte, pad)); err != nil {
					return err
				}
				written += pad
			}
		}
		if _, err := w.out.Write(orderedBodies[i]); err != nil {
			return err
		}
		written += e.Size
	}
	return nil
}

// ContentHash returns the sealed content hash computed by Finish.
// Returns the zero Digest for checkpoint profiles.
func (w *Writer) ContentHash() Digest {
	return w.contentHash
}

// WriteToBytes is a convenience that writes to an in-memory buffer and returns
// the full file bytes.
func WriteToBytes(profile Profile, version Version, sections []SectionInput, opts ...WriterOption) ([]byte, error) {
	var buf bytes.Buffer
	wr := NewWriter(&buf, profile, version, opts...)
	for _, s := range sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
