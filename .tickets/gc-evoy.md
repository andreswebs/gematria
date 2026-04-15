---
id: gc-evoy
status: closed
deps: [gc-a4ob]
links: []
created: 2026-04-15T19:25:18Z
type: task
priority: 2
assignee: Andre Silva
parent: gc-43xt
tags: [transliteration, agents-doc, task]
---
# Update AGENTS.md with transliteration documentation

Update `AGENTS.md` to document the transliteration feature for agent consumers. Purely additive — no existing content removed.

## Sections to add or update

### Recommended Flags table (extend)

Add rows:
- `Transliterate Latin input as Hebrew words` → `--transliterate` / `-t`
- `Select transliteration scheme` → `--scheme academic` (default) or `--scheme israeli`
- `All valid --scheme values` → `academic`, `israeli`

### Input Formats section (new subsection)

Add a \"Word Transliteration\" subsection between \"Latin Transliteration\" and \"Mixed Input\":

> When `--transliterate` / `-t` is set, Latin input is interpreted as full Hebrew words (phonetic transliteration), not as letter aliases. The scheme is selected by `--scheme`:
>
> - `academic` (default): strict consonantal mapping, ASCII fallbacks for diacritics
> - `israeli`: modern Israeli phonetic with matres lectionis (vowels mapped to ו/י/א/ה)
>
> ```sh
> gematria -t shalom                       # Compute שלום (academic)
> gematria -t --scheme israeli gadol       # Compute גדול (israeli)
> ```
>
> Without `-t`, Latin tokens are letter aliases (existing behavior). The two modes do not overlap.

### Output Schemas (extend Compute result)

Update the Compute result example to show `scheme` field with `omitempty` note:

```json
{
  \"input\": \"shalom\",
  \"system\": \"hechrachi\",
  \"scheme\": \"academic\",
  \"total\": 376,
  \"letters\": [ ... ]
}
```

Note: `scheme` is omitted when transliteration was not used.

### JSON error schema (extend with UnknownWordError case)

Add an example for `UnknownWordError`:

```json
{
  \"error\": \"input \\\"qzxw\\\" cannot be transliterated in scheme \\\"academic\\\"\",
  \"invalid_input\": \"qzxw\",
  \"scheme\": \"academic\",
  \"position\": 0,
  \"suggestions\": []
}
```

### Environment Variables table (extend)

Add row: `GEMATRIA_SCHEME` → `--scheme` → `Default transliteration scheme; lazy validation (only when -t is active)`.

### Quick Start (extend)

Add at least 2 examples using `-t`.

## Spec References

- docs/specs/transliteration.md (full feature spec)
- See parent epic.

## Acceptance Criteria

- [ ] Recommended Flags table includes `--transliterate`/`-t` and `--scheme`
- [ ] Input Formats has a new \"Word Transliteration\" subsection
- [ ] Compute result JSON schema example shows `scheme` field
- [ ] `scheme` documented as `omitempty` in JSON output
- [ ] JSON error schema has UnknownWordError example with `scheme` field
- [ ] Environment Variables table includes `GEMATRIA_SCHEME` with lazy-validation note
- [ ] Quick Start has at least 2 transliteration examples
- [ ] No existing AGENTS.md content removed

