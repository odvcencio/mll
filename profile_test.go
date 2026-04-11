package mll

import "testing"

func TestProfileRequirements(t *testing.T) {
	// Sealed must not contain OPTM.
	if IsForbidden(ProfileSealed, TagOPTM) != true {
		t.Error("sealed should forbid OPTM")
	}
	// Checkpoint must not contain SGNM.
	if IsForbidden(ProfileCheckpoint, TagSGNM) != true {
		t.Error("checkpoint should forbid SGNM")
	}
	// Weights-only must not contain KRNL, PLAN, ENTR, BUFF, OPTM.
	for _, tag := range [][4]byte{TagKRNL, TagPLAN, TagENTR, TagBUFF, TagOPTM} {
		if !IsForbidden(ProfileWeightsOnly, tag) {
			t.Errorf("weights-only should forbid %v", tag)
		}
	}
	// HEAD, STRG, PARM, TNSR are required across all profiles.
	for _, tag := range [][4]byte{TagHEAD, TagSTRG, TagPARM, TagTNSR} {
		for _, p := range []Profile{ProfileSealed, ProfileCheckpoint, ProfileWeightsOnly} {
			if !IsRequired(p, tag) {
				t.Errorf("%v should be required for profile %d", tag, p)
			}
		}
	}
	// OPTM is required in checkpoint.
	if !IsRequired(ProfileCheckpoint, TagOPTM) {
		t.Error("checkpoint should require OPTM")
	}
}

func TestCustomTagsAreNotForbidden(t *testing.T) {
	custom := [4]byte{'X', 'B', 'A', 'R'}
	for _, profile := range []Profile{ProfileSealed, ProfileCheckpoint, ProfileWeightsOnly} {
		if IsForbidden(profile, custom) {
			t.Fatalf("custom tag should not be forbidden for profile %d", profile)
		}
		if IsRequired(profile, custom) {
			t.Fatalf("custom tag should not be required for profile %d", profile)
		}
	}
}
