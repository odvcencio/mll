package mll

import (
	"bytes"
	"testing"
)

func TestPlanSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	b := NewPlanBuilder()
	b.Add(PlanStep{
		EntryRef: Ref{Tag: TagENTR, Index: 0},
		Kind:     PlanStepKernel,
		NameIdx:  strg.Intern("step0"),
	})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadPlanSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Steps) != 1 || decoded.Steps[0].Kind != PlanStepKernel {
		t.Fatalf("got %+v", decoded)
	}
}
