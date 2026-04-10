package mll

import (
	"bytes"
	"fmt"
	"io"
)

// PlanStepKind identifies one plan step variant.
const (
	PlanStepKernel uint8 = 0
	PlanStepHostOp uint8 = 1
)

// PlanStep is one step in a PLAN section.
type PlanStep struct {
	EntryRef  Ref
	Kind      uint8
	NameIdx   uint32
	KernelRef Ref // valid when Kind == PlanStepKernel
	Inputs    []Ref
	Outputs   []Ref
}

type PlanBuilder struct {
	steps []PlanStep
}

func NewPlanBuilder() *PlanBuilder { return &PlanBuilder{} }

func (b *PlanBuilder) Add(s PlanStep) { b.steps = append(b.steps, s) }

// Write layout: u32 count + repeat{ Ref entry, u8 kind, u32 name, Ref kernel,
//
//	u32 in_count, Ref[in], u32 out_count, Ref[out] }
func (b *PlanBuilder) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(b.steps))); err != nil {
		return err
	}
	for _, s := range b.steps {
		if _, err := w.Write(s.EntryRef.Encode()); err != nil {
			return err
		}
		if _, err := w.Write([]byte{s.Kind}); err != nil {
			return err
		}
		if err := WriteUint32LE(w, s.NameIdx); err != nil {
			return err
		}
		if _, err := w.Write(s.KernelRef.Encode()); err != nil {
			return err
		}
		if err := writeRefSlice(w, s.Inputs); err != nil {
			return err
		}
		if err := writeRefSlice(w, s.Outputs); err != nil {
			return err
		}
	}
	return nil
}

func writeRefSlice(w io.Writer, refs []Ref) error {
	if err := WriteUint32LE(w, uint32(len(refs))); err != nil {
		return err
	}
	for _, r := range refs {
		if _, err := w.Write(r.Encode()); err != nil {
			return err
		}
	}
	return nil
}

type PlanSection struct {
	Steps []PlanStep
}

func ReadPlanSection(data []byte) (PlanSection, error) {
	r := bytes.NewReader(data)
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return PlanSection{}, fmt.Errorf("mll: PLAN count: %w", err)
	}
	count, _ := ReadUint32LE(cBuf)
	s := PlanSection{Steps: make([]PlanStep, count)}
	for i := uint32(0); i < count; i++ {
		erBuf, err := readBytes(r, 8)
		if err != nil {
			return PlanSection{}, err
		}
		er, _ := DecodeRef(erBuf)
		s.Steps[i].EntryRef = er
		kBuf, err := readBytes(r, 1)
		if err != nil {
			return PlanSection{}, err
		}
		s.Steps[i].Kind = kBuf[0]
		nBuf, err := readBytes(r, 4)
		if err != nil {
			return PlanSection{}, err
		}
		s.Steps[i].NameIdx, _ = ReadUint32LE(nBuf)
		krBuf, err := readBytes(r, 8)
		if err != nil {
			return PlanSection{}, err
		}
		kr, _ := DecodeRef(krBuf)
		s.Steps[i].KernelRef = kr
		inputs, err := readRefSlice(r)
		if err != nil {
			return PlanSection{}, err
		}
		s.Steps[i].Inputs = inputs
		outputs, err := readRefSlice(r)
		if err != nil {
			return PlanSection{}, err
		}
		s.Steps[i].Outputs = outputs
	}
	return s, nil
}

func readRefSlice(r *bytes.Reader) ([]Ref, error) {
	cBuf, err := readBytes(r, 4)
	if err != nil {
		return nil, err
	}
	count, _ := ReadUint32LE(cBuf)
	out := make([]Ref, count)
	for i := uint32(0); i < count; i++ {
		refBuf, err := readBytes(r, 8)
		if err != nil {
			return nil, err
		}
		ref, err := DecodeRef(refBuf)
		if err != nil {
			return nil, err
		}
		out[i] = ref
	}
	return out, nil
}
