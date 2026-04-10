package mll

// sectionRequirement describes whether a section is required, optional, or forbidden per profile.
type sectionRequirement int

const (
	requirementForbidden sectionRequirement = iota
	requirementOptional
	requirementRequired
)

// profileRules maps (profile, section tag) → requirement. Derived from spec §Profiles.
// This table is the single source of truth for profile validation.
var profileRules = map[Profile]map[[4]byte]sectionRequirement{
	ProfileSealed: {
		TagHEAD: requirementRequired,
		TagSTRG: requirementRequired,
		TagENUM: requirementOptional,
		TagDIMS: requirementRequired,
		TagTYPE: requirementOptional,
		TagPARM: requirementRequired,
		TagENTR: requirementRequired,
		TagBUFF: requirementOptional,
		TagKRNL: requirementOptional,
		TagPLAN: requirementOptional,
		TagMEMP: requirementOptional,
		TagTNSR: requirementRequired,
		TagOPTM: requirementForbidden,
		TagSCHM: requirementOptional,
		TagSGNM: requirementOptional,
	},
	ProfileCheckpoint: {
		TagHEAD: requirementRequired,
		TagSTRG: requirementRequired,
		TagENUM: requirementOptional,
		TagDIMS: requirementRequired,
		TagTYPE: requirementOptional,
		TagPARM: requirementRequired,
		TagENTR: requirementRequired,
		TagBUFF: requirementOptional,
		TagKRNL: requirementOptional,
		TagPLAN: requirementOptional,
		TagMEMP: requirementOptional,
		TagTNSR: requirementRequired,
		TagOPTM: requirementRequired,
		TagSCHM: requirementOptional,
		TagSGNM: requirementForbidden,
	},
	ProfileWeightsOnly: {
		TagHEAD: requirementRequired,
		TagSTRG: requirementRequired,
		TagENUM: requirementOptional,
		TagDIMS: requirementOptional,
		TagTYPE: requirementOptional,
		TagPARM: requirementRequired,
		TagENTR: requirementForbidden,
		TagBUFF: requirementForbidden,
		TagKRNL: requirementForbidden,
		TagPLAN: requirementForbidden,
		TagMEMP: requirementOptional,
		TagTNSR: requirementRequired,
		TagOPTM: requirementForbidden,
		TagSCHM: requirementOptional,
		TagSGNM: requirementOptional,
	},
}

// IsRequired reports whether a section tag is required for the given profile.
func IsRequired(p Profile, tag [4]byte) bool {
	rules, ok := profileRules[p]
	if !ok {
		return false
	}
	return rules[tag] == requirementRequired
}

// IsForbidden reports whether a section tag is forbidden for the given profile.
func IsForbidden(p Profile, tag [4]byte) bool {
	rules, ok := profileRules[p]
	if !ok {
		return false
	}
	return rules[tag] == requirementForbidden
}
