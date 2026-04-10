package mll

import "testing"

func TestMagicBytes(t *testing.T) {
	want := [4]byte{'M', 'L', 'L', 0}
	if Magic != want {
		t.Fatalf("magic: got %v, want %v", Magic, want)
	}
}

func TestVersion10(t *testing.T) {
	if V1_0.Major != 1 || V1_0.Minor != 0 {
		t.Fatalf("V1_0: got %+v", V1_0)
	}
	if V1_0.Uint16() != 0x0100 {
		t.Fatalf("V1_0 uint16: got %x, want 0x0100", V1_0.Uint16())
	}
}

func TestProfileBytes(t *testing.T) {
	if ProfileSealed != 0x01 {
		t.Fatalf("ProfileSealed: got %x", ProfileSealed)
	}
	if ProfileCheckpoint != 0x02 {
		t.Fatalf("ProfileCheckpoint: got %x", ProfileCheckpoint)
	}
	if ProfileWeightsOnly != 0x03 {
		t.Fatalf("ProfileWeightsOnly: got %x", ProfileWeightsOnly)
	}
}

func TestHeaderSize(t *testing.T) {
	if HeaderSize != 24 {
		t.Fatalf("HeaderSize: got %d, want 24", HeaderSize)
	}
}

func TestDirectoryEntrySize(t *testing.T) {
	if DirectoryEntrySize != 64 {
		t.Fatalf("DirectoryEntrySize: got %d, want 64", DirectoryEntrySize)
	}
}
