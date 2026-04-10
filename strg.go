package mll

import (
	"bytes"
	"fmt"
	"io"
	"sort"
)

// StringTable represents the STRG section: a list of interned strings
// referenced by u32 index from elsewhere in the file.
type StringTable struct {
	strings []string
	index   map[string]uint32
}

// NewStringTableBuilder returns an empty string table builder.
func NewStringTableBuilder() *StringTable {
	return &StringTable{
		index: make(map[string]uint32),
	}
}

// Intern returns the index for s, inserting it if not already present.
// During accumulation, indices are assigned in first-seen order.
// CanonicalizeLexicographic rewrites indices to canonical order.
func (t *StringTable) Intern(s string) uint32 {
	if idx, ok := t.index[s]; ok {
		return idx
	}
	idx := uint32(len(t.strings))
	t.strings = append(t.strings, s)
	t.index[s] = idx
	return idx
}

// Size returns the number of strings in the table.
func (t *StringTable) Size() int {
	return len(t.strings)
}

// At returns the string at the given index.
func (t *StringTable) At(idx uint32) string {
	if int(idx) >= len(t.strings) {
		return ""
	}
	return t.strings[idx]
}

// Lookup returns the index of s, or (0, false) if not present.
func (t *StringTable) Lookup(s string) (uint32, bool) {
	idx, ok := t.index[s]
	return idx, ok
}

// Contains reports whether s is in the table.
func (t *StringTable) Contains(s string) bool {
	_, ok := t.index[s]
	return ok
}

// CanonicalizeLexicographic re-sorts the strings into lexicographic UTF-8 byte
// order and returns the remapping table: remap[old_index] = new_index.
// Callers must rewrite every string reference in every section using this map.
func (t *StringTable) CanonicalizeLexicographic() map[uint32]uint32 {
	if len(t.strings) == 0 {
		return nil
	}
	type indexed struct {
		str string
		old uint32
	}
	items := make([]indexed, len(t.strings))
	for i, s := range t.strings {
		items[i] = indexed{str: s, old: uint32(i)}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].str < items[j].str
	})
	remap := make(map[uint32]uint32, len(items))
	newStrings := make([]string, len(items))
	newIndex := make(map[string]uint32, len(items))
	for i, item := range items {
		newIdx := uint32(i)
		remap[item.old] = newIdx
		newStrings[i] = item.str
		newIndex[item.str] = newIdx
	}
	t.strings = newStrings
	t.index = newIndex
	return remap
}

// Write encodes the string table to w as a STRG section body.
// Format:
//
//	u32 string_count
//	for each string:
//	  u32 length
//	  utf8 bytes
func (t *StringTable) Write(w io.Writer) error {
	if err := WriteUint32LE(w, uint32(len(t.strings))); err != nil {
		return err
	}
	for _, s := range t.strings {
		if err := WriteUint32LE(w, uint32(len(s))); err != nil {
			return err
		}
		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
	}
	return nil
}

// ReadStringTable parses a STRG section body from b.
func ReadStringTable(b []byte) (*StringTable, error) {
	if len(b) < 4 {
		return nil, fmt.Errorf("mll: STRG needs at least 4 bytes")
	}
	r := bytes.NewReader(b)
	var count uint32
	var countBuf [4]byte
	if _, err := io.ReadFull(r, countBuf[:]); err != nil {
		return nil, err
	}
	count, _ = ReadUint32LE(countBuf[:])
	t := NewStringTableBuilder()
	for i := uint32(0); i < count; i++ {
		var lenBuf [4]byte
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			return nil, fmt.Errorf("mll: STRG string %d length: %w", i, err)
		}
		length, _ := ReadUint32LE(lenBuf[:])
		strBuf := make([]byte, length)
		if _, err := io.ReadFull(r, strBuf); err != nil {
			return nil, fmt.Errorf("mll: STRG string %d body: %w", i, err)
		}
		t.Intern(string(strBuf))
	}
	return t, nil
}
