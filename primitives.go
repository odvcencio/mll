package mll

// Primitive kinds in the MLL abstract data model.
type PrimitiveKind uint8

const (
	KindNull PrimitiveKind = iota
	KindBool
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat16
	KindFloat32
	KindFloat64
	KindString
	KindBytes
	KindList
	KindMap
	KindTensor
	KindEnum
	KindRef
	KindDim
	KindValue
)

// Typed scalar primitives. Go types directly model the MLL abstract data model.
type (
	Bool    bool
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
)

// Float16 stores a 16-bit IEEE 754 half-precision float as uint16 bits.
// Use F16FromFloat32 / F16ToFloat32 for conversions.
type Float16 uint16

// String is an interned UTF-8 string. At the data model level it's a Go string;
// binary encoding replaces it with a u32 index into the STRG section.
type String string

// Bytes is a raw byte blob with no encoding assumptions.
type Bytes []byte
