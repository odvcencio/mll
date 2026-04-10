package mll

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVectorMinimalReproduces(t *testing.T) {
	artifactPath := filepath.Join("testdata", "v1", "minimal.mllb")
	hashPath := filepath.Join("testdata", "v1", "minimal.hash")

	artifactBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	hashBytes, err := os.ReadFile(hashPath)
	if err != nil {
		t.Fatalf("read hash: %v", err)
	}
	expectedHex := strings.TrimSpace(string(hashBytes))
	if _, err := hex.DecodeString(expectedHex); err != nil {
		t.Fatalf("parse hash: %v", err)
	}

	r, err := ReadBytes(artifactBytes, WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	entries := r.DirectoryEntries()
	got := SealedContentHash(r.Version(), r.Profile(), 0, entries)
	if hex.EncodeToString(got[:]) != expectedHex {
		t.Fatalf("hash mismatch:\n got  %s\n want %s", hex.EncodeToString(got[:]), expectedHex)
	}
}

func TestVectorTinyEmbedReproduces(t *testing.T) {
	artifactPath := filepath.Join("testdata", "v1", "tiny_embed.mllb")
	hashPath := filepath.Join("testdata", "v1", "tiny_embed.hash")

	artifactBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	hashBytes, err := os.ReadFile(hashPath)
	if err != nil {
		t.Fatalf("read hash: %v", err)
	}
	expectedHex := strings.TrimSpace(string(hashBytes))
	if _, err := hex.DecodeString(expectedHex); err != nil {
		t.Fatalf("parse hash: %v", err)
	}

	r, err := ReadBytes(artifactBytes, WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if r.Profile() != ProfileSealed {
		t.Errorf("profile: got %d", r.Profile())
	}

	entries := r.DirectoryEntries()
	got := SealedContentHash(r.Version(), r.Profile(), 0, entries)
	if hex.EncodeToString(got[:]) != expectedHex {
		t.Fatalf("hash mismatch:\n got  %s\n want %s", hex.EncodeToString(got[:]), expectedHex)
	}

	// Sanity: the file should contain all sealed-required sections.
	requiredTags := [][4]byte{TagHEAD, TagSTRG, TagDIMS, TagPARM, TagENTR, TagTNSR}
	present := make(map[[4]byte]bool)
	for _, e := range entries {
		present[e.Tag] = true
	}
	for _, tag := range requiredTags {
		if !present[tag] {
			t.Errorf("missing required section %v in tiny_embed test vector", tag)
		}
	}
}
