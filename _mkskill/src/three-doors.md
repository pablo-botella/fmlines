---
mkskill:
  pos: 30
  in: readme
---

## The three doors

All of them speak `linereader.IoLineReader` (anything with a
`ReadLine() ([]byte, error)` — `github.com/pablo-botella/linereader` is the
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

