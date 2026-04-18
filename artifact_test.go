package mll

import "testing"

func TestSealedArtifactBuilderMarshalValidates(t *testing.T) {
	a := NewSealedArtifact("demo")
	a.AddDim("D", 2)
	if err := a.AddTensor("weights", DTypeF32, []uint64{2}, make([]byte, 8)); err != nil {
		t.Fatal(err)
	}

	data, hash, err := a.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if hash == (Digest{}) {
		t.Fatal("content hash is zero")
	}

	r, err := ReadBytes(data, WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	got, err := r.ContentHash()
	if err != nil {
		t.Fatal(err)
	}
	if got != hash {
		t.Fatalf("content hash mismatch: got %x want %x", got, hash)
	}
}

func TestSealedArtifactBuilderRejectsWrongTensorSize(t *testing.T) {
	a := NewSealedArtifact("demo")
	if err := a.AddTensor("weights", DTypeF32, []uint64{2}, make([]byte, 7)); err == nil {
		t.Fatal("expected tensor size error")
	}
}
