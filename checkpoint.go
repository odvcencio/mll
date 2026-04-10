package mll

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// CheckpointOptions configures a checkpoint writer.
type CheckpointOptions struct {
	// SlackBytes is the number of padding bytes reserved at the end of each
	// rewritable section (OPTM, TNSR). Reserved for a v1.x in-place rewrite
	// optimization; v1.0 always does full-rewrite-and-rename.
	SlackBytes uint64

	// SkipRequirementCheck disables the profile required-section check on
	// the underlying Writer. Intended for unit tests that exercise
	// checkpoint save/rewrite bookkeeping without constructing every
	// section type the checkpoint profile requires. Production callers
	// should leave this false.
	SkipRequirementCheck bool
}

// Checkpoint is a mutable MLL file of profile ProfileCheckpoint. Subsequent
// Save() calls rewrite via full sibling file + rename in v1.0.
//
// Generation counter increments on every save and is stored in HEAD.Generation.
// Callers set the HEAD section via SetSection with any Generation value they
// like; Save() is responsible for re-encoding HEAD with the current generation
// counter before writing, so on-disk and in-memory values always agree.
type Checkpoint struct {
	path       string
	opts       CheckpointOptions
	sections   map[[4]byte]SectionInput
	generation uint64
}

// NewCheckpoint creates a new checkpoint writer that will emit to path.
// If path exists, opens it and reads the existing generation.
func NewCheckpoint(path string, opts CheckpointOptions) (*Checkpoint, error) {
	ckpt := &Checkpoint{
		path:     path,
		opts:     opts,
		sections: make(map[[4]byte]SectionInput),
	}
	if data, err := os.ReadFile(path); err == nil {
		r, err := ReadBytes(data)
		if err != nil {
			return nil, fmt.Errorf("mll: existing checkpoint %s: %w", path, err)
		}
		if r.Profile() != ProfileCheckpoint {
			return nil, fmt.Errorf("mll: file %s has profile %d, want checkpoint", path, r.Profile())
		}
		if headBytes, ok := r.Section(TagHEAD); ok {
			head, err := ReadHeadSection(headBytes)
			if err != nil {
				return nil, err
			}
			ckpt.generation = head.Generation
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return ckpt, nil
}

// SetSection stores a section to be written on the next Save. The HEAD
// section's Generation field is overwritten by Save(); callers do not need
// to track it manually.
func (c *Checkpoint) SetSection(s SectionInput) {
	c.sections[s.Tag] = s
}

// Generation returns the current generation counter (monotonically incremented by Save).
func (c *Checkpoint) Generation() uint64 {
	return c.generation
}

// Save writes the checkpoint to disk via atomic sibling-file-plus-rename.
// Bumps the generation counter and re-encodes HEAD with the new value so
// on-disk HEAD.Generation matches Generation(). The in-memory counter is
// only advanced after the atomic rename succeeds.
func (c *Checkpoint) Save() error {
	nextGen := c.generation + 1

	// Re-encode HEAD with the new generation.
	headSec, ok := c.sections[TagHEAD]
	if !ok {
		return fmt.Errorf("mll: checkpoint save: HEAD section not set")
	}
	head, err := ReadHeadSection(headSec.Body)
	if err != nil {
		return fmt.Errorf("mll: checkpoint save: re-read HEAD: %w", err)
	}
	head.Generation = nextGen
	var headBuf bytes.Buffer
	if err := head.Write(&headBuf); err != nil {
		return fmt.Errorf("mll: checkpoint save: re-encode HEAD: %w", err)
	}
	headSec.Body = headBuf.Bytes()
	headSec.DigestBody = nil // checkpoint HEAD hashes full body
	c.sections[TagHEAD] = headSec

	dir := filepath.Dir(c.path)
	tmpPath := filepath.Join(dir, filepath.Base(c.path)+".tmp")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	var writerOpts []WriterOption
	if c.opts.SkipRequirementCheck {
		writerOpts = append(writerOpts, WithSkipRequirementCheck())
	}
	wr := NewWriter(tmpFile, ProfileCheckpoint, V1_0, writerOpts...)
	for _, s := range c.sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		return err
	}
	c.generation = nextGen
	return nil
}

// Close releases any resources held by the checkpoint. In v1.0 this is a no-op
// because Save closes its own file handle.
func (c *Checkpoint) Close() error {
	return nil
}
