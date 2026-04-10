package mll

import "testing"

func TestTensorMetadata(t *testing.T) {
	tn := Tensor{
		Name:   "token_embedding",
		DType:  DTypeF16,
		Shape:  []Dimension{DimSymbol("V"), DimSymbol("D")},
		Layout: LayoutRowMajor,
	}
	if tn.Name != "token_embedding" {
		t.Fatalf("name mismatch")
	}
	if tn.DType != DTypeF16 {
		t.Fatalf("dtype mismatch")
	}
	if len(tn.Shape) != 2 {
		t.Fatalf("shape rank: got %d", len(tn.Shape))
	}
}
