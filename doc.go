// Package mll implements the v1.0 Machine Learning Language binary artifact
// format.
//
// MLL files are sectioned binary containers for machine learning artifacts.
// The package provides the fixed file header and directory codecs, profile
// validation, canonical ordering and content hashing, section-specific
// encoders/decoders, a reader, a writer, and checkpoint save support.
package mll
