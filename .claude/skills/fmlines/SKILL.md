---
name: fmlines
description: "Front matter line extractor for Go — tokenizes headers line by line (fences, key-values, sections, invalids) over ot4go/linereader, leaving the document body unread in the reader. Use when working with github.com/ot4go/fmlines or when a Go project needs to read .md front matter or bare .fm headers without a YAML engine."
---

# fmlines — agent notes

Go package `github.com/ot4go/fmlines`: a front matter LINE EXTRACTOR — not
a parser. It tokenizes and classifies; interpreting the dialect (which keys
mean what) is the consumer's job. Depends only on
`github.com/ot4go/linereader`.

## API

```go
var lines fmlines.FmLines
kind, err := lines.ReadLine(lr)  // one line in (appended), its FmLineType out
err = lines.ReadHeader(lr)       // fenced .md header; STOPS after closing ---
err = lines.ReadAll(lr)          // bare .fm: every line to EOF

// lr is any linereader.IoLineReader: interface{ ReadLine() ([]byte, error) }

flr := fmlines.NewFmLineReader(r io.Reader, lines *FmLines) // nil lines → creates one
flr.ReadLine() / flr.ReadHeader() / flr.ReadAll()           // same, no params
```

## The token (FmLine)

- `Type`: Invalid=0, StartBlock=0x0001, EndBlock=0x0002, KeyValue=0x0010,
  Section=0x0020, Empty=0x0100 (bitmask-style values).
- `RawLine` is VERBATIM — the single source of truth. `Name()`/`Value()`
  are raw slices of it (`[:KvSplit]` / `[KvSplit+1:]`): NO trimming, NO
  unquoting — deliberately (cleaning is interpretation).
- `KvSplit` = the FIRST ':' (values may carry more colons). `Left` =
  indentation width.
- Tree doubly linked at birth: `Parent`, `Children *FmLines` (nil on
  leaves), `Level`. An indented line hangs from its innermost section;
  walking back, the first line further left decides (a section adopts,
  anything else = no parent; Empty and fences don't count).

## Semantics that matter

- ReadHeader consumes NOTHING beyond the closing fence: the body stays in
  the reader (that positioning is why linereader is underneath). Chain
  pattern: ReadHeader, then keep using the same reader for the body.
- No opening fence = no front matter: the one consumed line comes back as
  the only token — read is read, nothing gets lost. Unclosed fence = error.
- Unrecognized lines (YAML lists `- x`, multiline scalars…) classify as
  Invalid but keep RawLine, LineNum and Parent: recoverable by any consumer
  that wants to parse further. Deliberately NO array support, NO ToMap —
  both were tried and removed: extraction only.
- Scope: a minimalist take for SIMPLE front matter. Needing more control
  (real YAML: lists, types, the full spec) is the signal to use a full
  parser (gopkg.in/yaml.v3, goccy/go-yaml) — do not grow that machinery
  here.

## Tests

`test/fmlines/` — 10 tests: fenced/bare/no-fence/unclosed/empty, Hugo-style
lists surviving as adopted Invalids, the doubly linked tree, deep nesting,
and the FmLineReader bundle. `go test ./...`.
