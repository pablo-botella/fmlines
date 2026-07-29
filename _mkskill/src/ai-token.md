---
mkskill:
  pos: 220
  in: ai*
---

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

