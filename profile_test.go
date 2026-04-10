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
