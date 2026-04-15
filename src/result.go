package gematria

// LetterResult is a single-letter entry in a Compute result. It embeds the
// full Letter dictionary entry (name, meaning, position, aliases) and adds
// the computed numeric Value for the selected gematria system.
//
// The embedding promotes all Letter fields directly onto LetterResult so that
// output formatters can access Name, Meaning, Position, and Char without an
// extra dereference, while Value holds the system-specific computed number.
type LetterResult struct {
	Letter
	Value int
}

// Result is the return value of Compute. It is self-contained: all fields
// needed by any output formatter (line, value, card, json) are present without
// requiring additional dictionary lookups.
//
// Scheme is the empty string for results produced by Compute and
// ComputeFromLetters. ComputeTransliterated populates it with the scheme used
// for the Latin→Hebrew resolution. JSON output uses omitempty so the field is
// absent in non-transliterated results.
type Result struct {
	Input   string         // original input string as provided by the caller
	System  System         // gematria system used for computation
	Scheme  Scheme         // transliteration scheme used (empty when not transliterated)
	Total   int            // sum of all per-letter Values
	Letters []LetterResult // per-letter breakdown in input order
}
