package mll

// ValueKind is the discriminator for a Value primitive.
type ValueKind uint8

const (
	ValueKindInvalid       ValueKind = 0
	ValueKindTensor        ValueKind = 1
	ValueKindKVCache       ValueKind = 2
	ValueKindCandidatePack ValueKind = 3
)

// TensorType describes a tensor's type (element type and shape).
type TensorType struct {
	DType DType
	Shape []Dimension
}

// KVCacheType describes a transformer key-value cache.
type KVCacheType struct {
	Layers  int
	Heads   int
	HeadDim int
}

// CandidatePackType describes a retrieval candidate pack.
type CandidatePackType struct {
	Rank int // 2 or 3
}

// ValueType is a discriminated union over the three first-class value kinds.
type ValueType struct {
	Kind          ValueKind
	Tensor        *TensorType
	KVCache       *KVCacheType
	CandidatePack *CandidatePackType
}

// ValueOfTensor constructs a tensor-kind ValueType.
func ValueOfTensor(t TensorType) ValueType {
	return ValueType{Kind: ValueKindTensor, Tensor: &t}
}

// ValueOfKVCache constructs a kv_cache-kind ValueType.
func ValueOfKVCache(k KVCacheType) ValueType {
	return ValueType{Kind: ValueKindKVCache, KVCache: &k}
}

// ValueOfCandidatePack constructs a candidate_pack-kind ValueType.
func ValueOfCandidatePack(c CandidatePackType) ValueType {
	return ValueType{Kind: ValueKindCandidatePack, CandidatePack: &c}
}
