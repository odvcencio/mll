package mll

// CustomChunk carries an X*** section tag and opaque body bytes.
type CustomChunk struct {
	Tag  [4]byte
	Body []byte
}
