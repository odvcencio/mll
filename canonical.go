package mll

import (
	"bytes"
	"sort"
)

// coreSectionOrder is the fixed canonical order for core sections in sealed
// and weights-only profiles.
var coreSectionOrder = [][4]byte{
	TagHEAD,
	TagSTRG,
	TagENUM,
	TagDIMS,
	TagTYPE,
	TagPARM,
	TagENTR,
	TagBUFF,
	TagKRNL,
	TagPLAN,
	TagMEMP,
	TagTNSR,
	TagSCHM,
	// custom chunks (X*) go here, sorted lexicographically
	TagSGNM,
}

// coreTagRank returns the canonical position of a core tag, or -1 if not core.
func coreTagRank(tag [4]byte) int {
	for i, core := range coreSectionOrder {
		if core == tag {
			return i
		}
	}
	return -1
}

// CanonicalSectionOrder returns the directory entries in canonical order
// for sealed and weights-only profiles. Core sections appear in the fixed
// order defined by coreSectionOrder; custom chunks (X*) sort lexicographically
// after SCHM and before SGNM. Checkpoint profile returns the input unchanged.
func CanonicalSectionOrder(entries []DirectoryEntry, profile Profile) []DirectoryEntry {
	if profile == ProfileCheckpoint {
		out := make([]DirectoryEntry, len(entries))
		copy(out, entries)
		return out
	}
	// Split into core, custom, and sgnm buckets.
	var cores []DirectoryEntry
	var customs []DirectoryEntry
	var sgnms []DirectoryEntry
	for _, e := range entries {
		switch {
		case e.Tag == TagSGNM:
			sgnms = append(sgnms, e)
		case IsCustomTag(e.Tag):
			customs = append(customs, e)
		default:
			cores = append(cores, e)
		}
	}
	// Sort cores by fixed rank.
	sort.SliceStable(cores, func(i, j int) bool {
		return coreTagRank(cores[i].Tag) < coreTagRank(cores[j].Tag)
	})
	// Sort customs lexicographically.
	sort.SliceStable(customs, func(i, j int) bool {
		return bytes.Compare(customs[i].Tag[:], customs[j].Tag[:]) < 0
	})
	out := make([]DirectoryEntry, 0, len(entries))
	out = append(out, cores...)
	out = append(out, customs...)
	out = append(out, sgnms...)
	return out
}
