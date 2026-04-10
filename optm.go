package mll

import (
	"bytes"
	"fmt"
	"io"
)

// Optimizer kinds for OPTM.
const (
	OptimizerAdamW uint8 = 0
	OptimizerSGD   uint8 = 1
	OptimizerLAMB  uint8 = 2
)

// OptmSection is the OPTM section (checkpoint-only).
type OptmSection struct {
	Kind          uint8
	Step          uint64
	LRSchedule    uint8
	LRStateBytes  []byte
	Generation    uint64
	MomentTensors []Ref
}

// OptmBuilder accumulates optimizer state.
type OptmBuilder struct {
	section OptmSection
}

// NewOptmBuilder returns a builder seeded with the given optimizer kind.
func NewOptmBuilder(kind uint8) *OptmBuilder {
	return &OptmBuilder{section: OptmSection{Kind: kind}}
}

// SetStep sets the current optimizer step.
func (b *OptmBuilder) SetStep(step uint64) { b.section.Step = step }

// SetGeneration sets the checkpoint generation this OPTM entry belongs to.
func (b *OptmBuilder) SetGeneration(gen uint64) { b.section.Generation = gen }

// AddMomentTensor records a reference to a moment tensor in TNSR.
func (b *OptmBuilder) AddMomentTensor(ref Ref) {
	b.section.MomentTensors = append(b.section.MomentTensors, ref)
}

// Write encodes the OPTM section body.
// Layout: u8 kind + u64 step + u8 lr_schedule + u32 lr_state_len + lr_state + u64 generation + u32 moment_count + Ref[moment_count]
func (b *OptmBuilder) Write(w io.Writer) error {
	s := b.section
	if _, err := w.Write([]byte{s.Kind}); err != nil {
		return err
	}
	if err := WriteUint64LE(w, s.Step); err != nil {
		return err
	}
	if _, err := w.Write([]byte{s.LRSchedule}); err != nil {
		return err
	}
	if err := WriteUint32LE(w, uint32(len(s.LRStateBytes))); err != nil {
		return err
	}
	if len(s.LRStateBytes) > 0 {
		if _, err := w.Write(s.LRStateBytes); err != nil {
			return err
		}
	}
	if err := WriteUint64LE(w, s.Generation); err != nil {
		return err
	}
	if err := WriteUint32LE(w, uint32(len(s.MomentTensors))); err != nil {
		return err
	}
	for _, r := range s.MomentTensors {
		if _, err := w.Write(r.Encode()); err != nil {
			return err
		}
	}
	return nil
}

// ReadOptmSection decodes an OPTM section body.
func ReadOptmSection(data []byte) (OptmSection, error) {
	r := bytes.NewReader(data)
	kBuf, err := readBytes(r, 1)
	if err != nil {
		return OptmSection{}, fmt.Errorf("mll: OPTM kind: %w", err)
	}
	var s OptmSection
	s.Kind = kBuf[0]
	stBuf, err := readBytes(r, 8)
	if err != nil {
		return OptmSection{}, err
	}
	s.Step, _ = ReadUint64LE(stBuf)
	lrKBuf, err := readBytes(r, 1)
	if err != nil {
		return OptmSection{}, err
	}
	s.LRSchedule = lrKBuf[0]
	lrLBuf, err := readBytes(r, 4)
	if err != nil {
		return OptmSection{}, err
	}
	lrLen, _ := ReadUint32LE(lrLBuf)
	if lrLen > 0 {
		s.LRStateBytes, err = readBytes(r, int(lrLen))
		if err != nil {
			return OptmSection{}, err
		}
	}
	gBuf, err := readBytes(r, 8)
	if err != nil {
		return OptmSection{}, err
	}
	s.Generation, _ = ReadUint64LE(gBuf)
	mcBuf, err := readBytes(r, 4)
	if err != nil {
		return OptmSection{}, err
	}
	momentCount, _ := ReadUint32LE(mcBuf)
	s.MomentTensors = make([]Ref, momentCount)
	for i := uint32(0); i < momentCount; i++ {
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return OptmSection{}, err
		}
		ref, _ := DecodeRef(refBuf)
		s.MomentTensors[i] = ref
	}
	return s, nil
}
