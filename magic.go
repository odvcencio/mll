package mll

// Magic is the four-byte identifier at the start of every MLL binary file.
var Magic = [4]byte{'M', 'L', 'L', 0}

// HeaderSize is the fixed size of the file header in bytes.
const HeaderSize = 24

// DirectoryEntrySize is the fixed size of one section directory entry in bytes.
const DirectoryEntrySize = 64

// Version describes an MLL format version.
type Version struct {
	Major uint8
	Minor uint8
}

// Uint16 returns the 16-bit encoded form (major in high byte, minor in low byte).
func (v Version) Uint16() uint16 {
	return uint16(v.Major)<<8 | uint16(v.Minor)
}

// V1_0 is MLL format version 1.0.
var V1_0 = Version{Major: 1, Minor: 0}

// Profile identifies the role of an MLL artifact.
type Profile uint8

const (
	ProfileSealed      Profile = 0x01
	ProfileCheckpoint  Profile = 0x02
	ProfileWeightsOnly Profile = 0x03
)

// FileFlag values.
const (
	FileFlagHasSignature uint8 = 1 << 0
)

// SectionFlag values (u16 in directory entries).
const (
	SectionFlagRequired   uint16 = 1 << 0
	SectionFlagSkippable  uint16 = 1 << 1
	SectionFlagExternal   uint16 = 1 << 2
	SectionFlagCompressed uint16 = 1 << 3
	SectionFlagAligned    uint16 = 1 << 4
	SectionFlagSchemaless uint16 = 1 << 5
)

// Section tags for the core set.
var (
	TagHEAD = [4]byte{'H', 'E', 'A', 'D'}
	TagSTRG = [4]byte{'S', 'T', 'R', 'G'}
	TagENUM = [4]byte{'E', 'N', 'U', 'M'}
	TagDIMS = [4]byte{'D', 'I', 'M', 'S'}
	TagTYPE = [4]byte{'T', 'Y', 'P', 'E'}
	TagPARM = [4]byte{'P', 'A', 'R', 'M'}
	TagENTR = [4]byte{'E', 'N', 'T', 'R'}
	TagBUFF = [4]byte{'B', 'U', 'F', 'F'}
	TagKRNL = [4]byte{'K', 'R', 'N', 'L'}
	TagPLAN = [4]byte{'P', 'L', 'A', 'N'}
	TagMEMP = [4]byte{'M', 'E', 'M', 'P'}
	TagTNSR = [4]byte{'T', 'N', 'S', 'R'}
	TagOPTM = [4]byte{'O', 'P', 'T', 'M'}
	TagSCHM = [4]byte{'S', 'C', 'H', 'M'}
	TagSGNM = [4]byte{'S', 'G', 'N', 'M'}
)

// IsCustomTag reports whether a tag is in the custom chunk tag space (X***).
func IsCustomTag(tag [4]byte) bool {
	return tag[0] == 'X'
}
