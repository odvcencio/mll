package mll

import (
	"bytes"
	"testing"
)

func TestOptmSectionRoundTrip(t *testing.T) {
	b := NewOptmBuilder(OptimizerAdamW)
	b.SetStep(1000)
	b.SetGeneration(3)
	b.AddMomentTensor(Ref{Tag: TagTNSR, Index: 0})
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadOptmSection(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Kind != OptimizerAdamW || decoded.Step != 1000 || decoded.Generation != 3 {
		t.Fatalf("got %+v", decoded)
	}
	if len(decoded.MomentTensors) != 1 {
		t.Fatalf("moments: got %d", len(decoded.MomentTensors))
	}
}
