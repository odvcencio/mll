package mll

import "testing"

func TestValueTensor(t *testing.T) {
	v := ValueOfTensor(TensorType{DType: DTypeF16, Shape: []Dimension{DimSymbol("T")}})
	if v.Kind != ValueKindTensor {
		t.Fatalf("kind: got %d", v.Kind)
	}
	if v.Tensor == nil {
		t.Fatal("tensor should not be nil")
	}
}

func TestValueKVCache(t *testing.T) {
	v := ValueOfKVCache(KVCacheType{Layers: 6, Heads: 6, HeadDim: 64})
	if v.Kind != ValueKindKVCache {
		t.Fatalf("kind: got %d", v.Kind)
	}
}

func TestValueCandidatePack(t *testing.T) {
	v := ValueOfCandidatePack(CandidatePackType{Rank: 2})
	if v.Kind != ValueKindCandidatePack {
		t.Fatalf("kind: got %d", v.Kind)
	}
}
