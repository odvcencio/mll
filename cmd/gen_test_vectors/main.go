package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/odvcencio/mll"
)

func main() {
	outDir := "testdata/v1"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := genMinimal(outDir); err != nil {
		log.Fatalf("minimal: %v", err)
	}
	if err := genTinyEmbed(outDir); err != nil {
		log.Fatalf("tiny_embed: %v", err)
	}
	if err := genWeightsOnly(outDir); err != nil {
		log.Fatalf("weights_only: %v", err)
	}
	if err := genCheckpointGeneration(outDir); err != nil {
		log.Fatalf("checkpoint_generation: %v", err)
	}
	if err := genSignedEd25519(outDir); err != nil {
		log.Fatalf("signed_ed25519: %v", err)
	}
	if err := genCorruptDigest(outDir); err != nil {
		log.Fatalf("corrupt_digest: %v", err)
	}
	if err := genBadRef(outDir); err != nil {
		log.Fatalf("bad_ref: %v", err)
	}
	fmt.Println("wrote test vectors")
}

// genMinimal writes a minimal sealed artifact containing only HEAD + STRG.
// Uses WithSkipRequirementCheck to isolate canonicalization testing from
// section-specific machinery.
func genMinimal(dir string) error {
	strg := mll.NewStringTableBuilder()
	strg.Intern("minimal")
	strg.CanonicalizeLexicographic()
	nameIdx, _ := strg.Lookup("minimal")

	head := mll.HeadSection{Name: nameIdx}
	var headBuf bytes.Buffer
	head.Write(&headBuf)

	var strgBuf bytes.Buffer
	strg.Write(&strgBuf)

	sections := []mll.SectionInput{
		{
			Tag:           mll.TagHEAD,
			Body:          headBuf.Bytes(),
			DigestBody:    head.DigestBody(mll.ProfileSealed),
			Flags:         mll.SectionFlagRequired,
			SchemaVersion: 1,
		},
		{Tag: mll.TagSTRG, Body: strgBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
	}

	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0, mll.WithSkipRequirementCheck())
	for _, s := range sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "minimal.mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	return os.WriteFile(filepath.Join(dir, "minimal.hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644)
}

// genTinyEmbed writes a representative sealed inference artifact with all
// required sections for a sealed file: HEAD, STRG, DIMS, TYPE, PARM, ENTR, TNSR.
func genTinyEmbed(dir string) error {
	strg := mll.NewStringTableBuilder()
	strg.Intern("tiny_embed")
	strg.Intern("token_embedding")
	strg.Intern("t_embed")
	strg.Intern("forward")
	strg.Intern("D")
	strg.CanonicalizeLexicographic()

	nameIdx, _ := strg.Lookup("tiny_embed")
	paramNameIdx, _ := strg.Lookup("token_embedding")
	typeNameIdx, _ := strg.Lookup("t_embed")
	entryNameIdx, _ := strg.Lookup("forward")
	dimNameIdx, _ := strg.Lookup("D")

	// DIMS: D = 384 (static).
	dims := mll.NewDimsBuilder()
	dims.Add(mll.DimDecl{NameIdx: dimNameIdx, Bound: mll.DimBoundStatic, Value: 384})

	// TYPE: one tensor type shape=[D].
	dDim := mll.DimSymbol("D")
	dDim.SymbolIdx = dimNameIdx
	tys := mll.NewTypeBuilder()
	tys.AddTensorType(typeNameIdx, mll.DTypeF32, []mll.Dimension{dDim})

	// PARM: one parameter token_embedding : t_embed.
	parm := mll.NewParmBuilder()
	parm.Add(mll.ParmDecl{
		NameIdx:   paramNameIdx,
		TypeRef:   mll.Ref{Tag: mll.TagTYPE, Index: 0},
		Trainable: false,
	})

	// ENTR: one pipeline entry point.
	entr := mll.NewEntrBuilder()
	entr.Add(mll.EntryPoint{NameIdx: entryNameIdx, Kind: mll.EntryKindPipeline})

	// TNSR: one f32 tensor of 384 elements (= 1536 bytes).
	tnsr := mll.NewTnsrBuilder()
	tnsr.Add(mll.TensorEntry{
		NameIdx: paramNameIdx,
		DType:   mll.DTypeF32,
		Shape:   []uint64{384},
		Data:    make([]byte, 384*4),
	})

	head := mll.HeadSection{Name: nameIdx}

	var headBuf, strgBuf, dimsBuf, typeBuf, parmBuf, entrBuf, tnsrBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	dims.Write(&dimsBuf)
	tys.Write(&typeBuf)
	parm.Write(&parmBuf)
	entr.Write(&entrBuf)
	tnsr.Write(&tnsrBuf)

	sections := []mll.SectionInput{
		{Tag: mll.TagHEAD, Body: headBuf.Bytes(), DigestBody: head.DigestBody(mll.ProfileSealed), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagSTRG, Body: strgBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagDIMS, Body: dimsBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTYPE, Body: typeBuf.Bytes(), Flags: 0, SchemaVersion: 1},
		{Tag: mll.TagPARM, Body: parmBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagENTR, Body: entrBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTNSR, Body: tnsrBuf.Bytes(), Flags: mll.SectionFlagRequired | mll.SectionFlagAligned, SchemaVersion: 1},
	}

	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0)
	for _, s := range sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "tiny_embed.mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	return os.WriteFile(filepath.Join(dir, "tiny_embed.hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644)
}

func genWeightsOnly(dir string) error {
	strg := mll.NewStringTableBuilder()
	strg.Intern("weights_only")
	strg.Intern("weight")
	strg.Intern("t_weight")
	strg.CanonicalizeLexicographic()

	nameIdx, _ := strg.Lookup("weights_only")
	paramNameIdx, _ := strg.Lookup("weight")
	typeNameIdx, _ := strg.Lookup("t_weight")

	types := mll.NewTypeBuilder()
	types.AddTensorType(typeNameIdx, mll.DTypeF32, []mll.Dimension{mll.DimLiteral(4)})

	parm := mll.NewParmBuilder()
	parm.Add(mll.ParmDecl{NameIdx: paramNameIdx, TypeRef: mll.Ref{Tag: mll.TagTYPE, Index: 0}})

	tnsr := mll.NewTnsrBuilder()
	tnsr.Add(mll.TensorEntry{NameIdx: paramNameIdx, DType: mll.DTypeF32, Shape: []uint64{4}, Data: make([]byte, 16)})

	head := mll.HeadSection{Name: nameIdx}
	sections, err := sectionsFromBodies(mll.ProfileWeightsOnly, head, map[[4]byte]sectionBody{
		mll.TagSTRG: {body: mustBody(strg), flags: mll.SectionFlagRequired},
		mll.TagTYPE: {body: mustBody(types)},
		mll.TagPARM: {body: mustBody(parm), flags: mll.SectionFlagRequired},
		mll.TagTNSR: {body: mustBody(tnsr), flags: mll.SectionFlagRequired | mll.SectionFlagAligned},
	})
	if err != nil {
		return err
	}
	return writeVector(dir, "weights_only", mll.ProfileWeightsOnly, sections)
}

func genCheckpointGeneration(dir string) error {
	path := filepath.Join(dir, "checkpoint_generation.mllb")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	strg := mll.NewStringTableBuilder()
	nameIdx := strg.Intern("checkpoint_generation")
	paramNameIdx := strg.Intern("w")
	typeNameIdx := strg.Intern("t_w")
	entryNameIdx := strg.Intern("forward")
	dimNameIdx := strg.Intern("D")

	head := mll.HeadSection{Name: nameIdx}
	dims := mll.NewDimsBuilder()
	dims.Add(mll.DimDecl{NameIdx: dimNameIdx, Bound: mll.DimBoundStatic, Value: 4})
	dim := mll.DimSymbol("D")
	dim.SymbolIdx = dimNameIdx
	types := mll.NewTypeBuilder()
	types.AddTensorType(typeNameIdx, mll.DTypeF32, []mll.Dimension{dim})
	parm := mll.NewParmBuilder()
	parm.Add(mll.ParmDecl{NameIdx: paramNameIdx, TypeRef: mll.Ref{Tag: mll.TagTYPE, Index: 0}, Trainable: true})
	entr := mll.NewEntrBuilder()
	entr.Add(mll.EntryPoint{NameIdx: entryNameIdx, Kind: mll.EntryKindPipeline})
	tnsr := mll.NewTnsrBuilder()
	tnsr.Add(mll.TensorEntry{NameIdx: paramNameIdx, DType: mll.DTypeF32, Shape: []uint64{4}, Data: make([]byte, 16)})
	optm := mll.NewOptmBuilder(mll.OptimizerAdamW)

	ckpt, err := mll.NewCheckpoint(path, mll.CheckpointOptions{})
	if err != nil {
		return err
	}
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagHEAD, Body: mustBody(head), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagSTRG, Body: mustBody(strg), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagDIMS, Body: mustBody(dims), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagTYPE, Body: mustBody(types), SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagPARM, Body: mustBody(parm), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagENTR, Body: mustBody(entr), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagTNSR, Body: mustBody(tnsr), Flags: mll.SectionFlagRequired | mll.SectionFlagAligned, SchemaVersion: 1})
	ckpt.SetSection(mll.SectionInput{Tag: mll.TagOPTM, Body: mustBody(optm), Flags: mll.SectionFlagRequired, SchemaVersion: 1})
	if err := ckpt.Save(); err != nil {
		return err
	}
	if err := ckpt.Save(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "checkpoint_generation.generation"), []byte("2\n"), 0644)
}

func genSignedEd25519(dir string) error {
	sections, keyIdx, err := signableSections("signed_ed25519")
	if err != nil {
		return err
	}
	var unsigned bytes.Buffer
	unsignedWriter := mll.NewWriter(&unsigned, mll.ProfileSealed, mll.V1_0)
	for _, section := range sections {
		unsignedWriter.AddSection(section)
	}
	if err := unsignedWriter.Finish(); err != nil {
		return err
	}

	seed := bytes.Repeat([]byte{7}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	sgnm, err := mll.NewEd25519SgnmSection(keyIdx, privateKey, unsignedWriter.ContentHash())
	if err != nil {
		return err
	}

	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0)
	wr.SetFileFlag(mll.FileFlagHasSignature)
	for _, section := range sections {
		wr.AddSection(section)
	}
	wr.AddSection(mll.SectionInput{Tag: mll.TagSGNM, Body: mustBody(sgnm), SchemaVersion: 1})
	if err := wr.Finish(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "signed_ed25519.mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	if err := os.WriteFile(filepath.Join(dir, "signed_ed25519.hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "signed_ed25519.pub"), []byte(hex.EncodeToString(publicKey)+"\n"), 0644)
}

func genCorruptDigest(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, "tiny_embed.mllb"))
	if err != nil {
		return err
	}
	data = append([]byte(nil), data...)
	data[len(data)-1] ^= 0xff
	return os.WriteFile(filepath.Join(dir, "corrupt_digest.mllb"), data, 0644)
}

func genBadRef(dir string) error {
	strg := mll.NewStringTableBuilder()
	nameIdx := strg.Intern("bad_ref")
	paramIdx := strg.Intern("w")
	entryIdx := strg.Intern("forward")
	head := mll.HeadSection{Name: nameIdx}
	dims := mll.NewDimsBuilder()
	types := mll.NewTypeBuilder()
	parm := mll.NewParmBuilder()
	parm.Add(mll.ParmDecl{NameIdx: paramIdx, TypeRef: mll.Ref{Tag: mll.TagTYPE, Index: 9}})
	entr := mll.NewEntrBuilder()
	entr.Add(mll.EntryPoint{NameIdx: entryIdx, Kind: mll.EntryKindPipeline})
	tnsr := mll.NewTnsrBuilder()

	sections := []mll.SectionInput{
		{Tag: mll.TagHEAD, Body: mustBody(head), DigestBody: head.DigestBody(mll.ProfileSealed), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagSTRG, Body: mustBody(strg), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagDIMS, Body: mustBody(dims), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTYPE, Body: mustBody(types), SchemaVersion: 1},
		{Tag: mll.TagPARM, Body: mustBody(parm), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagENTR, Body: mustBody(entr), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTNSR, Body: mustBody(tnsr), Flags: mll.SectionFlagRequired | mll.SectionFlagAligned, SchemaVersion: 1},
	}
	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0)
	for _, section := range sections {
		wr.AddSection(section)
	}
	if err := wr.Finish(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "bad_ref.mllb"), out.Bytes(), 0644)
}

type sectionBody struct {
	body  []byte
	flags uint16
}

func sectionsFromBodies(profile mll.Profile, head mll.HeadSection, bodies map[[4]byte]sectionBody) ([]mll.SectionInput, error) {
	headBody := mustBody(head)
	sections := []mll.SectionInput{
		{Tag: mll.TagHEAD, Body: headBody, DigestBody: head.DigestBody(profile), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
	}
	for tag, body := range bodies {
		sections = append(sections, mll.SectionInput{Tag: tag, Body: body.body, Flags: body.flags, SchemaVersion: 1})
	}
	return sections, nil
}

func signableSections(name string) ([]mll.SectionInput, uint32, error) {
	strg := mll.NewStringTableBuilder()
	nameIdx := strg.Intern(name)
	entryIdx := strg.Intern("forward")
	keyIdx := strg.Intern("test-key")
	head := mll.HeadSection{Name: nameIdx}
	dims := mll.NewDimsBuilder()
	types := mll.NewTypeBuilder()
	parm := mll.NewParmBuilder()
	entr := mll.NewEntrBuilder()
	entr.Add(mll.EntryPoint{NameIdx: entryIdx, Kind: mll.EntryKindPipeline})
	tnsr := mll.NewTnsrBuilder()
	return []mll.SectionInput{
		{Tag: mll.TagHEAD, Body: mustBody(head), DigestBody: head.DigestBody(mll.ProfileSealed), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagSTRG, Body: mustBody(strg), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagDIMS, Body: mustBody(dims), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTYPE, Body: mustBody(types), SchemaVersion: 1},
		{Tag: mll.TagPARM, Body: mustBody(parm), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagENTR, Body: mustBody(entr), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTNSR, Body: mustBody(tnsr), Flags: mll.SectionFlagRequired | mll.SectionFlagAligned, SchemaVersion: 1},
	}, keyIdx, nil
}

func writeVector(dir, name string, profile mll.Profile, sections []mll.SectionInput) error {
	var out bytes.Buffer
	wr := mll.NewWriter(&out, profile, mll.V1_0)
	for _, section := range sections {
		wr.AddSection(section)
	}
	if err := wr.Finish(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, name+".mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	return os.WriteFile(filepath.Join(dir, name+".hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644)
}

func mustBody(section interface{ Write(io.Writer) error }) []byte {
	var buf bytes.Buffer
	if err := section.Write(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
