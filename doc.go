// Package mll implements the v1.0 MLL binary artifact format.
//
// MLL files are sectioned binary containers for ML artifacts. The package
// provides the fixed file header and directory codecs, semantic validation,
// canonical ordering and content hashing, section-specific encoders/decoders,
// signature verification, readers, writers, simple sealed-artifact building,
// and checkpoint save support.
package mll
