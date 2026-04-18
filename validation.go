package mll

import (
	"errors"
	"fmt"
	"math"
)

// Validate checks semantic invariants that span sections.
//
// Basic parsing validates headers, directory bounds, unsupported storage flags,
// and optional section digests. Validate goes further: it checks profile
// requirements, canonical ordering, string-table indices, cross-section refs,
// tensor byte counts, and signature-section consistency.
func (r *Reader) Validate() error {
	var errs []error
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if r.header.TotalFileSize != uint64(len(r.data)) {
		add("mll: header total_file_size=%d, actual=%d", r.header.TotalFileSize, len(r.data))
	}
	if _, ok := profileRules[r.header.Profile]; !ok {
		add("mll: unknown profile %d", r.header.Profile)
	}
	if unknown := r.header.Flags &^ FileFlagHasSignature; unknown != 0 {
		add("mll: unknown file flags 0x%02x", unknown)
	}

	seen := make(map[[4]byte]int, len(r.directory))
	for i, e := range r.directory {
		if first, ok := seen[e.Tag]; ok {
			add("mll: duplicate section %s at directory entries %d and %d", FormatTag(e.Tag), first, i)
		}
		seen[e.Tag] = i
		if !IsCustomTag(e.Tag) && !isKnownSectionTag(e.Tag) {
			add("mll: unknown core section tag %s", FormatTag(e.Tag))
		}
		if unknown := e.Flags &^ knownSectionFlags; unknown != 0 {
			add("mll: section %s has unknown flags 0x%04x", FormatTag(e.Tag), unknown)
		}
		if e.Flags&SectionFlagRequired != 0 && e.Flags&SectionFlagSkippable != 0 {
			add("mll: section %s cannot be both REQUIRED and SKIPPABLE", FormatTag(e.Tag))
		}
		if IsForbidden(r.header.Profile, e.Tag) {
			add("mll: section %s is forbidden in %s profile", FormatTag(e.Tag), r.header.Profile)
		}
	}
	for tag := range profileRules[r.header.Profile] {
		_, ok := seen[tag]
		if IsRequired(r.header.Profile, tag) && !ok {
			add("mll: %s profile requires section %s", r.header.Profile, FormatTag(tag))
		}
	}
	if r.header.Profile == ProfileSealed || r.header.Profile == ProfileWeightsOnly {
		canonical := CanonicalSectionOrder(r.directory, r.header.Profile)
		for i := range canonical {
			if canonical[i].Tag != r.directory[i].Tag {
				add("mll: section order is not canonical at entry %d: got %s, want %s", i, FormatTag(r.directory[i].Tag), FormatTag(canonical[i].Tag))
				break
			}
		}
	}

	if r.header.Flags&FileFlagHasSignature != 0 {
		if _, ok := r.Section(TagSGNM); !ok {
			add("mll: HAS_SIGNATURE is set but SGNM is missing")
		}
	} else if _, ok := r.Section(TagSGNM); ok {
		add("mll: SGNM is present but HAS_SIGNATURE is not set")
	}

	sections := parsedSections{}
	if body, ok := r.Section(TagSTRG); ok {
		sections.strg, sections.hasSTRG = parseForValidation(&errs, "STRG", func() (*StringTable, error) {
			return ReadStringTable(body)
		})
	}
	if body, ok := r.Section(TagHEAD); ok {
		sections.head, sections.hasHEAD = parseForValidation(&errs, "HEAD", func() (HeadSection, error) {
			return ReadHeadSection(body)
		})
	}
	if body, ok := r.Section(TagENUM); ok {
		sections.enum, sections.hasENUM = parseForValidation(&errs, "ENUM", func() (EnumSection, error) {
			return ReadEnumSection(body)
		})
	}
	if body, ok := r.Section(TagDIMS); ok {
		sections.dims, sections.hasDIMS = parseForValidation(&errs, "DIMS", func() (DimsSection, error) {
			return ReadDimsSection(body)
		})
	}
	if body, ok := r.Section(TagTYPE); ok {
		sections.typ, sections.hasTYPE = parseForValidation(&errs, "TYPE", func() (TypeSection, error) {
			return ReadTypeSection(body)
		})
	}
	if body, ok := r.Section(TagPARM); ok {
		sections.parm, sections.hasPARM = parseForValidation(&errs, "PARM", func() (ParmSection, error) {
			return ReadParmSection(body)
		})
	}
	if body, ok := r.Section(TagENTR); ok {
		sections.entr, sections.hasENTR = parseForValidation(&errs, "ENTR", func() (EntrSection, error) {
			return ReadEntrSection(body)
		})
	}
	if body, ok := r.Section(TagBUFF); ok {
		sections.buff, sections.hasBUFF = parseForValidation(&errs, "BUFF", func() (BuffSection, error) {
			return ReadBuffSection(body)
		})
	}
	if body, ok := r.Section(TagKRNL); ok {
		sections.krnl, sections.hasKRNL = parseForValidation(&errs, "KRNL", func() (KrnlSection, error) {
			return ReadKrnlSection(body)
		})
	}
	if body, ok := r.Section(TagPLAN); ok {
		sections.plan, sections.hasPLAN = parseForValidation(&errs, "PLAN", func() (PlanSection, error) {
			return ReadPlanSection(body)
		})
	}
	if body, ok := r.Section(TagMEMP); ok {
		sections.memp, sections.hasMEMP = parseForValidation(&errs, "MEMP", func() (MempSection, error) {
			return ReadMempSection(body)
		})
	}
	if body, ok := r.Section(TagTNSR); ok {
		sections.tnsr, sections.hasTNSR = parseForValidation(&errs, "TNSR", func() (TnsrSection, error) {
			return ReadTnsrSection(body)
		})
	}
	if body, ok := r.Section(TagOPTM); ok {
		sections.optm, sections.hasOPTM = parseForValidation(&errs, "OPTM", func() (OptmSection, error) {
			return ReadOptmSection(body)
		})
	}
	if body, ok := r.Section(TagSCHM); ok {
		_, _ = parseForValidation(&errs, "SCHM", func() (SchmSection, error) {
			return ReadSchmSection(body)
		})
	}
	if body, ok := r.Section(TagSGNM); ok {
		sections.sgnm, sections.hasSGNM = parseForValidation(&errs, "SGNM", func() (SgnmSection, error) {
			return ReadSgnmSection(body)
		})
	}

	validateParsedSections(add, sections)
	return errors.Join(errs...)
}

func isKnownSectionTag(tag [4]byte) bool {
	switch tag {
	case TagHEAD, TagSTRG, TagENUM, TagDIMS, TagTYPE, TagPARM, TagENTR,
		TagBUFF, TagKRNL, TagPLAN, TagMEMP, TagTNSR, TagOPTM, TagSCHM, TagSGNM:
		return true
	default:
		return false
	}
}

const knownSectionFlags = SectionFlagRequired |
	SectionFlagSkippable |
	SectionFlagExternal |
	SectionFlagCompressed |
	SectionFlagAligned |
	SectionFlagSchemaless

type parsedSections struct {
	strg *StringTable

	head HeadSection
	enum EnumSection
	dims DimsSection
	typ  TypeSection
	parm ParmSection
	entr EntrSection
	buff BuffSection
	krnl KrnlSection
	plan PlanSection
	memp MempSection
	tnsr TnsrSection
	optm OptmSection
	sgnm SgnmSection

	hasSTRG bool
	hasHEAD bool
	hasENUM bool
	hasDIMS bool
	hasTYPE bool
	hasPARM bool
	hasENTR bool
	hasBUFF bool
	hasKRNL bool
	hasPLAN bool
	hasMEMP bool
	hasTNSR bool
	hasOPTM bool
	hasSGNM bool
}

func parseForValidation[T any](errs *[]error, tag string, fn func() (T, error)) (out T, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			*errs = append(*errs, fmt.Errorf("mll: parse %s: panic: %v", tag, r))
			ok = false
		}
	}()
	out, err := fn()
	if err != nil {
		*errs = append(*errs, fmt.Errorf("mll: parse %s: %w", tag, err))
		return out, false
	}
	return out, true
}

func validateParsedSections(add func(string, ...any), s parsedSections) {
	var strCount uint32
	if s.hasSTRG {
		strCount = uint32(s.strg.Size())
	}
	checkString := func(ctx string, idx uint32, optional bool) {
		if !s.hasSTRG {
			return
		}
		if optional && idx == 0 {
			return
		}
		if idx >= strCount {
			add("mll: %s string index %d out of range (strings=%d)", ctx, idx, strCount)
		}
	}

	counts := map[[4]byte]uint32{}
	if s.hasENUM {
		counts[TagENUM] = uint32(len(s.enum.Enums))
	}
	if s.hasDIMS {
		counts[TagDIMS] = uint32(len(s.dims.Decls))
	}
	if s.hasTYPE {
		counts[TagTYPE] = uint32(len(s.typ.Decls))
	}
	if s.hasPARM {
		counts[TagPARM] = uint32(len(s.parm.Decls))
	}
	if s.hasENTR {
		counts[TagENTR] = uint32(len(s.entr.Entries))
	}
	if s.hasBUFF {
		counts[TagBUFF] = uint32(len(s.buff.Decls))
	}
	if s.hasKRNL {
		counts[TagKRNL] = uint32(len(s.krnl.Decls))
	}
	if s.hasPLAN {
		counts[TagPLAN] = uint32(len(s.plan.Steps))
	}
	if s.hasMEMP {
		counts[TagMEMP] = uint32(len(s.memp.Entries))
	}
	if s.hasTNSR {
		counts[TagTNSR] = uint32(len(s.tnsr.Tensors))
	}

	checkRef := func(ctx string, ref Ref) {
		n, ok := counts[ref.Tag]
		if !ok {
			add("mll: %s references missing section %s", ctx, FormatTag(ref.Tag))
			return
		}
		if ref.Index >= n {
			add("mll: %s reference %s[%d] out of range (count=%d)", ctx, FormatTag(ref.Tag), ref.Index, n)
		}
	}
	checkDim := func(ctx string, d Dimension) {}
	checkDim = func(ctx string, d Dimension) {
		switch d.Kind {
		case DimKindSymbol:
			if s.hasSTRG && d.SymbolIdx >= strCount {
				add("mll: %s dimension symbol index %d out of range (strings=%d)", ctx, d.SymbolIdx, strCount)
			}
		case DimKindExpr:
			if d.Expr != nil {
				checkDim(ctx+".left", d.Expr.Left)
				checkDim(ctx+".right", d.Expr.Right)
			}
		}
	}

	if s.hasHEAD {
		checkString("HEAD.name", s.head.Name, false)
		checkString("HEAD.description", s.head.Description, true)
		for i, m := range s.head.Metadata {
			checkString(fmt.Sprintf("HEAD.metadata[%d].key", i), m.Key, false)
			if m.Kind == HeadValueString {
				checkString(fmt.Sprintf("HEAD.metadata[%d].string", i), m.StringIdx, false)
			}
		}
	}
	if s.hasENUM {
		for i, e := range s.enum.Enums {
			checkString(fmt.Sprintf("ENUM[%d].name", i), e.Name, false)
			for j, v := range e.Values {
				checkString(fmt.Sprintf("ENUM[%d].values[%d]", i, j), v, false)
			}
		}
	}
	if s.hasDIMS {
		for i, d := range s.dims.Decls {
			checkString(fmt.Sprintf("DIMS[%d].name", i), d.NameIdx, false)
			if d.Bound != DimBoundDynamic && d.Bound != DimBoundStatic {
				add("mll: DIMS[%d] has invalid bound %d", i, d.Bound)
			}
		}
	}
	if s.hasTYPE {
		for i, d := range s.typ.Decls {
			checkString(fmt.Sprintf("TYPE[%d].name", i), d.NameIdx, false)
			for j, dim := range d.Shape {
				checkDim(fmt.Sprintf("TYPE[%d].shape[%d]", i, j), dim)
			}
		}
	}
	if s.hasPARM {
		for i, p := range s.parm.Decls {
			checkString(fmt.Sprintf("PARM[%d].name", i), p.NameIdx, false)
			checkString(fmt.Sprintf("PARM[%d].binding", i), p.BindingIdx, true)
			checkRef(fmt.Sprintf("PARM[%d].type", i), p.TypeRef)
		}
	}
	if s.hasENTR {
		for i, e := range s.entr.Entries {
			checkString(fmt.Sprintf("ENTR[%d].name", i), e.NameIdx, false)
			for j, b := range e.Inputs {
				checkString(fmt.Sprintf("ENTR[%d].inputs[%d].name", i, j), b.NameIdx, false)
				checkRef(fmt.Sprintf("ENTR[%d].inputs[%d].type", i, j), b.TypeRef)
			}
			for j, b := range e.Outputs {
				checkString(fmt.Sprintf("ENTR[%d].outputs[%d].name", i, j), b.NameIdx, false)
				checkRef(fmt.Sprintf("ENTR[%d].outputs[%d].type", i, j), b.TypeRef)
			}
		}
	}
	if s.hasBUFF {
		for i, b := range s.buff.Decls {
			checkString(fmt.Sprintf("BUFF[%d].name", i), b.NameIdx, false)
			checkRef(fmt.Sprintf("BUFF[%d].type", i), b.TypeRef)
		}
	}
	if s.hasKRNL {
		for i, k := range s.krnl.Decls {
			checkString(fmt.Sprintf("KRNL[%d].name", i), k.NameIdx, false)
		}
	}
	if s.hasPLAN {
		for i, p := range s.plan.Steps {
			checkRef(fmt.Sprintf("PLAN[%d].entry", i), p.EntryRef)
			checkString(fmt.Sprintf("PLAN[%d].name", i), p.NameIdx, false)
			if p.Kind == PlanStepKernel {
				checkRef(fmt.Sprintf("PLAN[%d].kernel", i), p.KernelRef)
			}
			for j, ref := range p.Inputs {
				checkRef(fmt.Sprintf("PLAN[%d].inputs[%d]", i, j), ref)
			}
			for j, ref := range p.Outputs {
				checkRef(fmt.Sprintf("PLAN[%d].outputs[%d]", i, j), ref)
			}
		}
	}
	if s.hasMEMP {
		for i, m := range s.memp.Entries {
			checkRef(fmt.Sprintf("MEMP[%d].param", i), m.ParamRef)
		}
	}
	if s.hasTNSR {
		for i, t := range s.tnsr.Tensors {
			checkString(fmt.Sprintf("TNSR[%d].name", i), t.NameIdx, false)
			validateTensorByteSize(add, i, t)
		}
	}
	if s.hasOPTM {
		for i, ref := range s.optm.MomentTensors {
			checkRef(fmt.Sprintf("OPTM.moments[%d]", i), ref)
		}
	}
	if s.hasSGNM {
		checkString("SGNM.key_id", s.sgnm.KeyIDIdx, false)
		switch s.sgnm.Algorithm {
		case SigAlgorithmNone:
			if len(s.sgnm.Signature) != 0 {
				add("mll: SGNM algorithm none must not carry signature bytes")
			}
		case SigAlgorithmEd25519:
			if len(s.sgnm.Signature) != 64 {
				add("mll: SGNM ed25519 signature length=%d, want 64", len(s.sgnm.Signature))
			}
		default:
			add("mll: SGNM has unsupported signature algorithm %d", s.sgnm.Algorithm)
		}
	}
}

func validateTensorByteSize(add func(string, ...any), idx int, t TensorEntry) {
	elemSize := t.DType.ElementSize()
	if elemSize == 0 {
		return
	}
	elements := uint64(1)
	for _, dim := range t.Shape {
		if dim != 0 && elements > math.MaxUint64/dim {
			add("mll: TNSR[%d] shape overflows uint64 element count", idx)
			return
		}
		elements *= dim
	}
	if elements > math.MaxUint64/uint64(elemSize) {
		add("mll: TNSR[%d] byte size overflows uint64", idx)
		return
	}
	want := elements * uint64(elemSize)
	if t.BodySize != want {
		add("mll: TNSR[%d] body_size=%d, want %d for dtype=%d shape=%v", idx, t.BodySize, want, t.DType, t.Shape)
	}
}
