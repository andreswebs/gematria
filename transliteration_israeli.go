package gematria

import (
	"fmt"
	"strings"
)

// israeliVowels is the set of single-byte vowel characters recognised by the
// israeli scheme. Unlike the academic scheme, vowels are not all dropped —
// they are position-sensitive (word-initial, medial, or word-final).
var israeliVowels = map[byte]bool{
	'a': true,
	'e': true,
	'i': true,
	'o': true,
	'u': true,
}

// israeliHasConsonantAfter returns true if any consonant appears in s strictly
// after position pos. It is used for word-final vowel detection: a vowel at
// pos is word-final when there is no consonant following it.
//
// Consonant tables are shared with the academic scheme (identical per spec §4.2).
func israeliHasConsonantAfter(s string, pos int) bool {
	for i := pos + 1; i < len(s); {
		// Check 2-char multi-char sequence first (greedy).
		if i+1 < len(s) {
			if _, ok := academicMultiChar[s[i:i+2]]; ok {
				return true
			}
		}
		c := s[i]
		// Check single consonant.
		if _, ok := academicSingleChar[c]; ok {
			return true
		}
		// Skip over vowels; errors on unknown chars are caught in the main loop.
		if israeliVowels[c] {
			i++
			continue
		}
		i++
	}
	return false
}

// transliterateIsraeli converts a single word-part (no spaces) from Latin to
// a slice of Hebrew runes using the israeli (modern phonetic) scheme.
//
// Consonant tables are identical to the academic scheme. Vowels are
// position-aware:
//   - a/e word-initial (before any letter emitted) → א (aleph vowel carrier)
//   - a/e medial                                   → dropped
//   - a/e word-final (no consonant follows)        → ה (he mater)
//   - i any position                               → י (yod mater)
//   - o/u word-initial                             → א + ו (aleph carrier + vav mater)
//   - o/u non-initial                              → ו (vav mater)
//
// Sofit substitution is applied after all letters are resolved (same rule as
// the academic scheme; uses sofitMap defined in transliteration_academic.go).
func transliterateIsraeli(input string) ([]rune, error) {
	s := strings.ToLower(input)
	var out []rune
	letterEmitted := false // true once any Hebrew letter has been appended

	for i := 0; i < len(s); {
		// Try 2-char multi-char consonant sequence first (greedy).
		if i+1 < len(s) {
			pair := s[i : i+2]
			if r, ok := academicMultiChar[pair]; ok {
				out = append(out, r)
				letterEmitted = true
				i += 2
				continue
			}
		}

		c := s[i]

		// Single consonant (same table as academic scheme).
		if r, ok := academicSingleChar[c]; ok {
			out = append(out, r)
			letterEmitted = true
			i++
			continue
		}

		// Vowel handling with position awareness.
		switch c {
		case 'a', 'e':
			if !letterEmitted {
				// Word-initial: aleph as vowel carrier.
				out = append(out, 'א')
				letterEmitted = true
			} else if !israeliHasConsonantAfter(s, i) {
				// Word-final: he as mater lectionis.
				out = append(out, 'ה')
			}
			// Medial: drop silently.
			i++

		case 'i':
			// Yod mater throughout.
			out = append(out, 'י')
			letterEmitted = true
			i++

		case 'o', 'u':
			if !letterEmitted {
				// Word-initial: aleph carrier + vav mater.
				out = append(out, 'א', 'ו')
				letterEmitted = true
			} else {
				// Non-initial: vav mater only.
				out = append(out, 'ו')
			}
			i++

		default:
			return nil, fmt.Errorf("israeli scheme: cannot map character %q in input %q", string(c), input)
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("israeli scheme: input %q produces no Hebrew letters", input)
	}

	// Apply sofit substitution to the last letter (same rule as academic).
	if sofit, ok := sofitMap[out[len(out)-1]]; ok {
		out[len(out)-1] = sofit
	}

	return out, nil
}
