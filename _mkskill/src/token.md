---
mkskill:
  pos: 20
  in: readme
---

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

