package mll

// CustomChunk carries an X*** section tag + opaque body bytes.
// Full MCD handling lands in Plan 2; Plan 1 just preserves the bytes.
type CustomChunk struct {
	Tag  [4]byte
	Body []byte
}
