---
mkskill:
  pos: 230
  in: ai*
---

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

