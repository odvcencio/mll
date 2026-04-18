package mll

import (
	"crypto/ed25519"
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

func TestVectorWeightsOnlyReproduces(t *testing.T) {
	r, expectedHex := readVectorWithHash(t, "weights_only")
	if r.Profile() != ProfileWeightsOnly {
		t.Errorf("profile: got %d", r.Profile())
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	got, err := r.ContentHash()
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(got[:]) != expectedHex {
		t.Fatalf("hash mismatch:\n got  %s\n want %s", hex.EncodeToString(got[:]), expectedHex)
	}
}

func TestVectorCheckpointGeneration(t *testing.T) {
	r, err := ReadFile(filepath.Join("testdata", "v1", "checkpoint_generation.mllb"), WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	headBytes, ok := r.Section(TagHEAD)
	if !ok {
		t.Fatal("missing HEAD")
	}
	head, err := ReadHeadSection(headBytes)
	if err != nil {
		t.Fatal(err)
	}
	wantBytes, err := os.ReadFile(filepath.Join("testdata", "v1", "checkpoint_generation.generation"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := head.Generation, strings.TrimSpace(string(wantBytes)); got != 2 || want != "2" {
		t.Fatalf("generation got %d, generation file %q", got, want)
	}
}

func TestVectorSignedEd25519(t *testing.T) {
	r, expectedHex := readVectorWithHash(t, "signed_ed25519")
	if err := r.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	got, err := r.ContentHash()
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(got[:]) != expectedHex {
		t.Fatalf("hash mismatch:\n got  %s\n want %s", hex.EncodeToString(got[:]), expectedHex)
	}
	pubHex, err := os.ReadFile(filepath.Join("testdata", "v1", "signed_ed25519.pub"))
	if err != nil {
		t.Fatal(err)
	}
	pub, err := hex.DecodeString(strings.TrimSpace(string(pubHex)))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.VerifySignature(ed25519.PublicKey(pub)); err != nil {
		t.Fatalf("verify signature: %v", err)
	}
}

func TestVectorCorruptDigestRejected(t *testing.T) {
	_, err := ReadFile(filepath.Join("testdata", "v1", "corrupt_digest.mllb"), WithDigestVerification())
	if err == nil {
		t.Fatal("expected corrupt digest vector to fail digest verification")
	}
}

func TestVectorBadRefValidationFails(t *testing.T) {
	r, err := ReadFile(filepath.Join("testdata", "v1", "bad_ref.mllb"), WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := r.Validate(); err == nil {
		t.Fatal("expected bad_ref vector to fail validation")
	}
}

func readVectorWithHash(t *testing.T, name string) (*Reader, string) {
	t.Helper()
	artifactPath := filepath.Join("testdata", "v1", name+".mllb")
	hashPath := filepath.Join("testdata", "v1", name+".hash")

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
	return r, expectedHex
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
