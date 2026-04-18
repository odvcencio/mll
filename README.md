# MLL: Machine Learning Language

MLL is a binary interchange format and Go implementation for machine learning
artifacts. It is designed to package model metadata, tensor declarations, tensor
bytes, entry points, memory planning hints, optimizer state, and signatures into
a deterministic sectioned file.

The name is pronounced "mill." It is `LLM` backwards on purpose.

This repository is the standalone v1.0 Go module for reading, writing, hashing,
and testing MLL binary artifacts.

## What Is Implemented

- Fixed 24-byte file header with `MLL\0` magic, version, profile, flags, file
  size, section count, and reader compatibility fields.
- 64-byte section directory entries with section tag, offset, size,
  BLAKE3-256 digest, flags, and schema version.
- Reader and writer APIs for in-memory bytes, files, section lookup, optional
  digest verification, profile validation, and canonical section ordering.
- Reproducible sealed and weights-only content hashes that ignore layout-only
  details such as offsets, padding, total file size, and signature bytes.
- Encoders and decoders for the core v1.0 sections: `HEAD`, `STRG`, `ENUM`,
  `DIMS`, `TYPE`, `PARM`, `ENTR`, `BUFF`, `KRNL`, `PLAN`, `MEMP`, `TNSR`,
  `OPTM`, `SCHM`, and `SGNM`.
- Checkpoint writer support with generation tracking and atomic
  sibling-file-plus-rename saves.
- Golden v1 test vectors for a minimal file and a representative sealed
  `tiny_embed` artifact.

## Format At A Glance

An MLL file is:

```text
24-byte header
64-byte directory entry * section_count
section bodies, optionally padded/aligned
```

Every section body is addressed by a four-byte tag. Core tags include:

| Tag | Purpose |
| --- | --- |
| `HEAD` | artifact name, description, timestamps, generation, capabilities, metadata |
| `STRG` | interned UTF-8 string table used by other sections |
| `DIMS` | symbolic and static dimension declarations |
| `TYPE` | tensor, KV-cache, and candidate-pack type declarations |
| `PARM` | model parameter declarations |
| `ENTR` | entry points for functions, pipelines, and kernels |
| `BUFF` | buffer declarations |
| `KRNL` | kernel declarations, currently opaque v1 payloads |
| `PLAN` | execution plan steps |
| `MEMP` | residency and access-count hints for weights |
| `TNSR` | tensor metadata plus raw tensor bytes |
| `OPTM` | checkpoint-only optimizer state |
| `SCHM` | schema extension section, empty in v1.0 |
| `SGNM` | signature metadata and signature bytes |

Custom extension sections use the `X***` tag space.

## Profiles

MLL v1.0 defines three artifact profiles:

| Profile | Intent | Notes |
| --- | --- | --- |
| `ProfileSealed` | Immutable inference artifact | Canonical section order, reproducible content hash, requires `HEAD`, `STRG`, `DIMS`, `PARM`, `ENTR`, and `TNSR`; forbids `OPTM`. |
| `ProfileCheckpoint` | Mutable training checkpoint | Preserves writer section order, requires `OPTM`, stores `HEAD.Generation`, and is saved through full rewrite plus atomic rename in v1.0. |
| `ProfileWeightsOnly` | Portable weights bundle | Canonical section order and reproducible content hash, requires `HEAD`, `STRG`, `PARM`, and `TNSR`; forbids execution and optimizer sections. |

The writer enforces required and forbidden sections by default. Tests and vector
generators can opt out with `WithSkipRequirementCheck()`.

## Install

```bash
go get github.com/odvcencio/mll
```

```go
import "github.com/odvcencio/mll"
```

## Read An Artifact

```go
r, err := mll.ReadFile("model.mllb", mll.WithDigestVerification())
if err != nil {
    return err
}

fmt.Printf("mll v%d.%d profile=%d sections=%d\n",
    r.Version().Major,
    r.Version().Minor,
    r.Profile(),
    r.SectionCount(),
)

if body, ok := r.Section(mll.TagTNSR); ok {
    tensors, err := mll.ReadTnsrSection(body)
    if err != nil {
        return err
    }
    fmt.Println("tensor count:", len(tensors.Tensors))
}
```

Digest verification is optional because callers sometimes need fast metadata
inspection. Use `WithDigestVerification()` when loading artifacts across a trust
boundary or before computing sealed content hashes.

## Write An Artifact

The writer composes complete files from encoded section bodies. Section-specific
builders produce those bodies.

```go
var out bytes.Buffer
wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0)

sections := []mll.SectionInput{
    {
        Tag:           mll.TagHEAD,
        Body:          headBody,
        DigestBody:    head.DigestBody(mll.ProfileSealed),
        Flags:         mll.SectionFlagRequired,
        SchemaVersion: 1,
    },
    {Tag: mll.TagSTRG, Body: stringTableBody, Flags: mll.SectionFlagRequired, SchemaVersion: 1},
    {Tag: mll.TagDIMS, Body: dimsBody, Flags: mll.SectionFlagRequired, SchemaVersion: 1},
    {Tag: mll.TagTYPE, Body: typeBody, SchemaVersion: 1},
    {Tag: mll.TagPARM, Body: parmBody, Flags: mll.SectionFlagRequired, SchemaVersion: 1},
    {Tag: mll.TagENTR, Body: entrBody, Flags: mll.SectionFlagRequired, SchemaVersion: 1},
    {
        Tag:           mll.TagTNSR,
        Body:          tensorBody,
        Flags:         mll.SectionFlagRequired | mll.SectionFlagAligned,
        SchemaVersion: 1,
    },
}

for _, section := range sections {
    wr.AddSection(section)
}

if err := wr.Finish(); err != nil {
    return err
}

contentHash := wr.ContentHash()
_ = contentHash
```

For a complete working writer example, see
[`cmd/gen_test_vectors/main.go`](cmd/gen_test_vectors/main.go). It builds the
`tiny_embed` sealed artifact with `HEAD`, `STRG`, `DIMS`, `TYPE`, `PARM`,
`ENTR`, and `TNSR`.

## Test Vectors

The v1 fixtures live in [`testdata/v1`](testdata/v1):

- `minimal.mllb` and `minimal.hash`: a small canonicalization fixture.
- `tiny_embed.mllb` and `tiny_embed.hash`: a representative sealed inference
  artifact with all required sealed sections.

Regenerate them with:

```bash
go run ./cmd/gen_test_vectors
```

Validate the module with:

```bash
go test ./...
```

## Current Boundaries

This package is the v1.0 binary core. The following behavior is intentional at
this stage:

- Reader support is limited to MLL major version 1.
- Sections marked `EXTERNAL` or `COMPRESSED` are rejected by the reader.
- `SGNM` stores signature metadata and bytes; cryptographic signature
  verification is not implemented here yet.
- `SCHM` is accepted as an empty section in v1.0.
- `KRNL` stores an opaque body so kernel payloads can round-trip while the full
  kernel DSL remains future work.
- Checkpoint saves use full file rewrite plus atomic rename; in-place updates
  are reserved for a future v1.x optimization.
- Quantized tensor byte accounting, especially `DTypeQ4`, is left to higher
  level code.

## Development

The canonical branch name for this repository is `main`.

Primary commands:

```bash
go test ./...
go run ./cmd/gen_test_vectors
```

The code is Apache-2.0 licensed. See [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
