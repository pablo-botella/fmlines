---
mkskill:
  pos: 40
  in: readme
---

## What it deliberately does not do

No YAML engine: no multiline lists (`- item` lines classify as `Invalid` —
but raw, numbered and adopted by their section, so nothing is lost), no
multiline strings, no anchors, no type coercion, no unquoting. When a tool
writes what a tool reads, that machinery is dead weight — and whoever needs
it has the raw lines, intact.

**Scope, honestly**: this is a minimalist take for parsing *simple* front
matter. When you need more control — real YAML, lists, types, the whole
specification — reach for a full parser such as `gopkg.in/yaml.v3` or
`github.com/goccy/go-yaml` instead; this package is not trying to compete
with them.

