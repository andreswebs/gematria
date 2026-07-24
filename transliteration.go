package gematria

import "strings"

// Scheme is the transliteration scheme to use when converting Latin input
// to Hebrew letters. It parallels the System type for gematria systems.
type Scheme string

const (
	// SchemeAcademic is the strict consonantal transliteration scheme.
	// Vowels in Latin input are dropped; output is biblical-style consonantal Hebrew.
	SchemeAcademic Scheme = "academic"

	// SchemeIsraeli is the modern Israeli phonetic transliteration scheme.
	// Vowels are mapped to matres lectionis (ו, י, א, ה).
	SchemeIsraeli Scheme = "israeli"
)

// ValidSchemes returns both supported transliteration schemes in stable order.
// Used in error messages and flag validation.
func ValidSchemes() []Scheme {
	return []Scheme{SchemeAcademic, SchemeIsraeli}
}

// Transliterate parses input as a Hebrew word or sequence of words using the
// given transliteration scheme, returning the resolved sequence of letters.
//
// The input is split on whitespace into word-parts. Each part is transliterated
// independently so that sofit substitution and word-initial vowel rules apply at
// part boundaries. The resulting letter sequences are concatenated in order.
//
// Word-parts that consist entirely of Hebrew Unicode characters (U+05D0–U+05EA)
// are passed through unchanged — each rune is resolved via LookupLetter.
// This allows Hebrew Unicode input to be used alongside Latin input.
//
// Returns *InvalidSchemeError if scheme is not recognised.
// Returns *UnknownWordError if a word-part cannot be resolved to Hebrew letters.
// The Position field of *UnknownWordError is the 0-based index of the failing
// word-part within the space-split token list.
func Transliterate(input string, scheme Scheme) ([]Letter, error) {
	switch scheme {
	case SchemeAcademic, SchemeIsraeli:
		// valid
	default:
		return nil, &InvalidSchemeError{Name: string(scheme), Valid: ValidSchemes()}
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, &UnknownWordError{Input: input, Scheme: scheme, Position: 0}
	}

	var result []Letter
	for i, part := range parts {
		ls, err := transliteratePart(part, scheme)
		if err != nil {
			return nil, &UnknownWordError{Input: part, Scheme: scheme, Position: i}
		}
		result = append(result, ls...)
	}
	return result, nil
}

// transliteratePart resolves a single space-free word-part to a Letter slice.
// All-Hebrew parts are passed through via LookupLetter. Latin parts are
// dispatched to the per-scheme transliterator.
func transliteratePart(part string, scheme Scheme) ([]Letter, error) {
	if isAllHebrew(part) {
		var result []Letter
		for _, r := range part {
			l, err := LookupLetter(r)
			if err != nil {
				return nil, err
			}
			result = append(result, l)
		}
		return result, nil
	}

	var runes []rune
	var err error
	switch scheme {
	case SchemeAcademic:
		runes, err = transliterateAcademic(part)
	case SchemeIsraeli:
		runes, err = transliterateIsraeli(part)
	}
	if err != nil {
		return nil, err
	}

	result := make([]Letter, 0, len(runes))
	for _, r := range runes {
		l, err := LookupLetter(r)
		if err != nil {
			return nil, err
		}
		result = append(result, l)
	}
	return result, nil
}

// isAllHebrew reports whether s is non-empty and consists entirely of Hebrew
// Unicode characters (U+05D0–U+05EA).
func isAllHebrew(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isHebrewRune(r) {
			return false
		}
	}
	return true
}
