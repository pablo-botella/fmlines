# fmlines

Minimalist front matter **line extractor** for Go — deliberately not a
parser: it tokenizes a header line by line and classifies each one; what the
lines *mean* (which keys, which sections — the dialect) is the consumer's
business. Every line read becomes a token, the unrecognized ones too — raw
and numbered, never lost.

```go
flr := fmlines.NewFmLineReader(file, nil)
if err := flr.ReadHeader(); err != nil {   // reads the --- fenced header, then STOPS
	return err
}
for _, line := range *flr.Lines {
	fmt.Println(line.Type, line.Name(), line.Value())
}
body, _ := flr.Reader.ReadLine() // the body is still in the reader, untouched
```

## The token

Each line comes back as an `FmLine`:

| Field | Meaning |
|---|---|
| `Type` | `StartBlock` / `EndBlock` (the `---` fences), `KeyValue`, `Section` (a bare `name:`), `Empty` (blank or `#` comment), `Invalid` (anything else — kept, never dropped) |
| `RawLine` | the line verbatim: the single source of truth |
| `LineNum` | 1-based position |
| `KvSplit` | index of the first `:` |
| `Left` | indentation width |
| `Parent` / `Children` / `Level` | the section tree, doubly linked at birth: an indented line hangs from its innermost section |

`Name()` and `Value()` are raw views over `RawLine` — no trimming, no
unquoting: cleaning up is interpretation, and interpretation belongs to the
consumer.

## The three doors

All of them speak `linereader.IoLineReader` (anything with a
`ReadLine() ([]byte, error)` — `github.com/ot4go/linereader` is the
reference):

- **`lines.ReadLine(lr)`** — one line in, its `FmLineType` out.
- **`lines.ReadHeader(lr)`** — a fenced header (`.md`): tokens from the
  opening `---` to the closing one, and **it stops right there** — the body
  stays in the reader, keep reading it there (by lines, by bytes, or hand
  the reader to any `io.Reader` consumer). No opening fence: no front
  matter — the one consumed line comes back as the only token. Unclosed
  fence: error.
- **`lines.ReadAll(lr)`** — a bare header (`.fm`): every line to EOF.

`FmLineReader` bundles a reader and its list for the common case:
`NewFmLineReader(r io.Reader, lines *FmLines)` (nil list: one is created),
then `flr.ReadLine()` / `flr.ReadHeader()` / `flr.ReadAll()` without
parameters.

## What it deliberately does not do

No YAML engine: no multiline lists (`- item` lines classify as `Invalid` —
but raw, numbered and adopted by their section, so nothing is lost), no
multiline strings, no anchors, no type coercion, no unquoting. When a tool
writes what a tool reads, that machinery is dead weight — and whoever needs
it has the raw lines, intact.

## Install

```
go get github.com/ot4go/fmlines
```

## License

MIT — see [LICENSE](LICENSE).
