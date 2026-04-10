package mll

import "testing"

func TestDimLiteral(t *testing.T) {
	d := DimLiteral(384)
	if d.Kind != DimKindLiteral {
		t.Fatalf("kind: got %d", d.Kind)
	}
	if d.Value != 384 {
		t.Fatalf("value: got %d", d.Value)
	}
}

func TestDimSymbol(t *testing.T) {
	d := DimSymbol("D")
	if d.Kind != DimKindSymbol {
		t.Fatalf("kind: got %d", d.Kind)
	}
	if d.Symbol != "D" {
		t.Fatalf("symbol: got %q", d.Symbol)
	}
}

func TestDimString(t *testing.T) {
	cases := []struct {
		d    Dimension
		want string
	}{
		{DimLiteral(384), "384"},
		{DimSymbol("D"), "D"},
	}
	for _, c := range cases {
		if got := c.d.String(); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
