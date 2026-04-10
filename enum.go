package mll

// EnumDecl declares a named enum type with its valid values.
type EnumDecl struct {
	Name   string
	Values []string
}

// HasValue reports whether the given value is valid for this enum.
func (e EnumDecl) HasValue(v string) bool {
	for _, candidate := range e.Values {
		if candidate == v {
			return true
		}
	}
	return false
}

// EnumValue is an instance of an enum: a type name plus a chosen value.
type EnumValue struct {
	Type  string // name of the enum type
	Value string // one of the declared values
}
