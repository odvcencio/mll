package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/mll"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "inspect":
		if err := runInspect(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  mll inspect [-verify-digests=true] <file.mllb>")
}

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	verifyDigests := fs.Bool("verify-digests", true, "verify every section digest while reading")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("mll inspect: expected exactly one file")
	}

	var opts []mll.ReadOption
	if *verifyDigests {
		opts = append(opts, mll.WithDigestVerification())
	}
	r, err := mll.ReadFile(fs.Arg(0), opts...)
	if err != nil {
		return err
	}

	h := r.Header()
	fmt.Printf("path: %s\n", fs.Arg(0))
	fmt.Printf("version: %s\n", h.Version)
	fmt.Printf("profile: %s\n", h.Profile)
	fmt.Printf("flags: 0x%02x%s\n", h.Flags, fileFlagSuffix(h.Flags))
	fmt.Printf("total_file_size: %d\n", h.TotalFileSize)
	fmt.Printf("section_count: %d\n", h.SectionCount)
	fmt.Printf("min_reader_minor: %d\n", h.MinReaderMinor)
	fmt.Printf("digest_verification: %s\n", boolStatus(*verifyDigests))
	if hash, err := r.ContentHash(); err == nil {
		fmt.Printf("content_hash: %s\n", hex.EncodeToString(hash[:]))
	}

	if err := r.Validate(); err != nil {
		fmt.Println("validation: failed")
		for _, line := range errorLines(err) {
			fmt.Printf("  - %s\n", line)
		}
	} else {
		fmt.Println("validation: ok")
	}

	fmt.Println("sections:")
	for i, e := range r.DirectoryEntries() {
		fmt.Printf("  - index: %d\n", i)
		fmt.Printf("    tag: %s\n", mll.FormatTag(e.Tag))
		fmt.Printf("    offset: %d\n", e.Offset)
		fmt.Printf("    size: %d\n", e.Size)
		fmt.Printf("    flags: 0x%04x%s\n", e.Flags, sectionFlagSuffix(e.Flags))
		fmt.Printf("    schema_version: %d\n", e.SchemaVersion)
		fmt.Printf("    digest: %s\n", hex.EncodeToString(e.Digest[:]))
	}
	return nil
}

func boolStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "skipped"
}

func fileFlagSuffix(flags uint8) string {
	var names []string
	if flags&mll.FileFlagHasSignature != 0 {
		names = append(names, "HAS_SIGNATURE")
	}
	if len(names) == 0 {
		return ""
	}
	return " (" + strings.Join(names, "|") + ")"
}

func sectionFlagSuffix(flags uint16) string {
	var names []string
	if flags&mll.SectionFlagRequired != 0 {
		names = append(names, "REQUIRED")
	}
	if flags&mll.SectionFlagSkippable != 0 {
		names = append(names, "SKIPPABLE")
	}
	if flags&mll.SectionFlagExternal != 0 {
		names = append(names, "EXTERNAL")
	}
	if flags&mll.SectionFlagCompressed != 0 {
		names = append(names, "COMPRESSED")
	}
	if flags&mll.SectionFlagAligned != 0 {
		names = append(names, "ALIGNED")
	}
	if flags&mll.SectionFlagSchemaless != 0 {
		names = append(names, "SCHEMALESS")
	}
	if len(names) == 0 {
		return ""
	}
	return " (" + strings.Join(names, "|") + ")"
}

func errorLines(err error) []string {
	type unwrapper interface {
		Unwrap() []error
	}
	if joined, ok := err.(unwrapper); ok {
		var out []string
		for _, child := range joined.Unwrap() {
			out = append(out, errorLines(child)...)
		}
		return out
	}
	return strings.Split(err.Error(), "\n")
}
