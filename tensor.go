package mll

// DType is the element type of a tensor.
type DType uint8

const (
	DTypeInvalid DType = 0
	DTypeI8      DType = 1
	DTypeI16     DType = 2
	DTypeI32     DType = 3
	DTypeI64     DType = 4
	DTypeU8      DType = 5
	DTypeU16     DType = 6
	DTypeU32     DType = 7
	DTypeU64     DType = 8
	DTypeF16     DType = 9
	DTypeF32     DType = 10
	DTypeF64     DType = 11
	DTypeQ4      DType = 12
	DTypeQ8      DType = 13
)

// Layout is the storage layout of a tensor's elements in memory.
type Layout uint8

const (
	LayoutRowMajor Layout = 0
	LayoutColMajor Layout = 1
)

// ElementSize returns the number of bytes per element for standard dtypes.
// Quantized types (Q4, Q8) require higher-level inspection to compute byte count.
func (d DType) ElementSize() int {
	switch d {
	case DTypeI8, DTypeU8, DTypeQ8:
		return 1
	case DTypeI16, DTypeU16, DTypeF16:
		return 2
	case DTypeI32, DTypeU32, DTypeF32:
		return 4
	case DTypeI64, DTypeU64, DTypeF64:
		return 8
	case DTypeQ4:
		return 0 // packed, requires dedicated accounting
	default:
		return 0
	}
}

// Tensor is a first-class primitive representing tensor metadata and a
// reference to its data. The raw bytes live in the TNSR section; this struct
// holds the metadata plus a pointer (by name) into TNSR.
type Tensor struct {
	Name       string      // name of the tensor; resolved to STRG index
	DType      DType       // element type
	Shape      []Dimension // shape as symbolic dimensions
	Layout     Layout      // memory layout
	DataOffset uint64      // byte offset within the TNSR section body (set during seal)
	DataSize   uint64      // byte length (set during seal)
}

// Dimension stub — replaced by dim.go in Task 1.6. Do not build on this; the
// stub exists only so tensor.go compiles until Task 1.6 lands.
type Dimension struct {
	Kind   uint8
	Symbol string
	Value  int64
}

// DimSymbol stub — replaced by the real constructor in dim.go in Task 1.6.
func DimSymbol(name string) Dimension {
	return Dimension{Kind: 1, Symbol: name}
}
