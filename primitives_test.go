package mll

import "testing"

func TestBool(t *testing.T) {
	v := Bool(true)
	if !v {
		t.Fatal("Bool(true) should be true")
	}
}

func TestInt(t *testing.T) {
	var i Int64 = -42
	if int64(i) != -42 {
		t.Fatalf("Int64: got %d", i)
	}
}

func TestFloat(t *testing.T) {
	var f Float32 = 3.14
	if float32(f) != 3.14 {
		t.Fatalf("Float32: got %v", f)
	}
}

func TestString(t *testing.T) {
	s := String("hello")
	if len(s) != 5 || s != "hello" {
		t.Fatalf("String: got %q", s)
	}
}

func TestBytes(t *testing.T) {
	b := Bytes{1, 2, 3}
	if len(b) != 3 || b[1] != 2 {
		t.Fatalf("Bytes: got %v", b)
	}
}
