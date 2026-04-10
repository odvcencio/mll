package mll

import (
	"fmt"
	"os"
	"sort"
)

// Reader represents a loaded MLL binary file.
type Reader struct {
	data       []byte
	header     FileHeader
	directory  []DirectoryEntry
	sectionMap map[[4]byte]int
}

// ReadOption configures Reader behavior.
type ReadOption func(*readerConfig)

type readerConfig struct {
	verifyDigests bool
}

// WithDigestVerification instructs the reader to verify every section's
// BLAKE3-256 digest against the directory entry.
func WithDigestVerification() ReadOption {
	return func(c *readerConfig) {
		c.verifyDigests = true
	}
}

// ReadBytes parses an MLL binary file from a byte slice.
// By default, section digests are NOT verified (for speed); pass
// WithDigestVerification() to enable verification.
func ReadBytes(data []byte, opts ...ReadOption) (*Reader, error) {
	cfg := readerConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("mll: file too small (%d < %d)", len(data), HeaderSize)
	}
	header, err := ReadHeader(data[:HeaderSize])
	if err != nil {
		return nil, fmt.Errorf("mll: header: %w", err)
	}
	if header.Version.Major != V1_0.Major {
		return nil, fmt.Errorf("mll: unsupported major version %d", header.Version.Major)
	}
	if header.MinReaderMinor > V1_0.Minor {
		return nil, fmt.Errorf("mll: file requires reader >= v1.%d", header.MinReaderMinor)
	}
	dirStart := HeaderSize
	dirEnd := dirStart + DirectoryEntrySize*int(header.SectionCount)
	if dirEnd > len(data) {
		return nil, fmt.Errorf("mll: directory truncated")
	}
	directory, err := ReadDirectory(data[dirStart:dirEnd], header.SectionCount)
	if err != nil {
		return nil, fmt.Errorf("mll: directory: %w", err)
	}
	sectionMap := make(map[[4]byte]int, len(directory))
	sortedByOffset := make([]int, len(directory))
	for i := range directory {
		sortedByOffset[i] = i
	}
	sort.SliceStable(sortedByOffset, func(i, j int) bool {
		return directory[sortedByOffset[i]].Offset < directory[sortedByOffset[j]].Offset
	})
	var prevEnd uint64 = uint64(HeaderSize) + uint64(len(directory))*DirectoryEntrySize
	for _, idx := range sortedByOffset {
		e := directory[idx]
		if e.Offset+e.Size > uint64(len(data)) {
			return nil, fmt.Errorf("mll: section %v out of bounds (offset=%d size=%d file=%d)", e.Tag, e.Offset, e.Size, len(data))
		}
		if e.Offset < prevEnd {
			return nil, fmt.Errorf("mll: section %v overlaps previous region (offset=%d, expected >= %d)", e.Tag, e.Offset, prevEnd)
		}
		prevEnd = e.Offset + e.Size
		if e.Flags&SectionFlagExternal != 0 {
			return nil, fmt.Errorf("mll: section %v has EXTERNAL flag, not supported in v1.0", e.Tag)
		}
		if e.Flags&SectionFlagCompressed != 0 {
			return nil, fmt.Errorf("mll: section %v has COMPRESSED flag, not supported in v1.0", e.Tag)
		}
		sectionMap[e.Tag] = idx
	}
	// Digest verification.
	//
	// Special case: under ProfileSealed and ProfileWeightsOnly, the HEAD
	// section's digest was computed by Writer over HeadSection.DigestBody(profile),
	// which zeroes created_unix_ms and generation so sealed content hashes are
	// reproducible across wall clocks. To verify, we must reproduce that same
	// transform here rather than hashing the raw on-disk body.
	if cfg.verifyDigests {
		for _, e := range directory {
			body := data[e.Offset : e.Offset+e.Size]
			digestInput := body
			if e.Tag == TagHEAD && (header.Profile == ProfileSealed || header.Profile == ProfileWeightsOnly) {
				head, err := ReadHeadSection(body)
				if err != nil {
					return nil, fmt.Errorf("mll: parse HEAD for digest verification: %w", err)
				}
				digestInput = head.DigestBody(header.Profile)
			}
			if HashBytes(digestInput) != e.Digest {
				return nil, fmt.Errorf("mll: digest mismatch for section %v", e.Tag)
			}
		}
	}
	return &Reader{
		data:       data,
		header:     header,
		directory:  directory,
		sectionMap: sectionMap,
	}, nil
}

// ReadFile is a convenience wrapper that reads from a file path.
func ReadFile(path string, opts ...ReadOption) (*Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadBytes(data, opts...)
}

// Version returns the file's format version.
func (r *Reader) Version() Version { return r.header.Version }

// Profile returns the file's profile byte.
func (r *Reader) Profile() Profile { return r.header.Profile }

// SectionCount returns the number of sections in the file.
func (r *Reader) SectionCount() uint32 { return r.header.SectionCount }

// DirectoryEntries returns a copy of the directory entries.
func (r *Reader) DirectoryEntries() []DirectoryEntry {
	out := make([]DirectoryEntry, len(r.directory))
	copy(out, r.directory)
	return out
}

// Section returns the body bytes for the section with the given tag.
// The returned slice is a view into the underlying file bytes; callers MUST
// NOT modify it. Returns (nil, false) if the section is not present.
func (r *Reader) Section(tag [4]byte) ([]byte, bool) {
	idx, ok := r.sectionMap[tag]
	if !ok {
		return nil, false
	}
	e := r.directory[idx]
	return r.data[e.Offset : e.Offset+e.Size], true
}
