---
mkskill:
  pos: 10
  in: readme
---

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

