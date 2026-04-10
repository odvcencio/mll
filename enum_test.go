package mll

import "testing"

func TestEnumDecl(t *testing.T) {
	e := EnumDecl{
		Name:   "BackendKind",
		Values: []string{"cuda", "metal"},
	}
	if e.Name != "BackendKind" {
		t.Fatal("name mismatch")
	}
	if !e.HasValue("cuda") {
		t.Fatal("should have cuda")
	}
	if e.HasValue("opencl") {
		t.Fatal("should not have opencl")
	}
}

func TestEnumValue(t *testing.T) {
	v := EnumValue{Type: "BackendKind", Value: "cuda"}
	if v.Type != "BackendKind" || v.Value != "cuda" {
		t.Fatalf("got %+v", v)
	}
}
