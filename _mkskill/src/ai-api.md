---
mkskill:
  pos: 210
  in: ai*
---

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

