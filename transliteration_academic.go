package gematria

import (
	"fmt"
	"strings"
)

// academicMultiChar maps 2-character Latin sequences to Hebrew runes.
// Checked before academicSingleChar (greedy longest-match-first).
var academicMultiChar = map[string]rune{
	"sh": 'ש', // Shin (300)
	"kh": 'ח', // Het (8)   — scholarly "kh" = guttural
	"ch": 'ח', // Het (8)   — synonym for "kh"
	"ts": 'צ', // Tsade (90)
	"tz": 'צ', // Tsade (90) — alternate ASCII spelling
	"ph": 'פ', // Pe (80)
}

// academicSingleChar maps single-byte Latin characters to Hebrew runes.
// Vowels (a,e,i,o,u) are absent — they are silently dropped.
var academicSingleChar = map[byte]rune{
	'\'': 'א', // Aleph (1)   — apostrophe = explicit glottal stop
	'b':  'ב', // Bet (2)
	'g':  'ג', // Gimel (3)
	'd':  'ד', // Dalet (4)
	'h':  'ה', // He (5)
	'v':  'ו', // Vav (6)
	'w':  'ו', // Vav (6)     — alternate ASCII spelling
	'z':  'ז', // Zayin (7)
	'x':  'ח', // Het (8)     — ASCII fallback for guttural (x for kh)
	'y':  'י', // Yod (10)
	'k':  'כ', // Kaf (20)
	'l':  'ל', // Lamed (30)
	'm':  'מ', // Mem (40)
	'n':  'נ', // Nun (50)
	's':  'ס', // Samekh (60) — plain sibilant; Shin requires "sh"
	'p':  'פ', // Pe (80)
	'f':  'פ', // Pe (80)     — fricative (peh without dagesh)
	'q':  'ק', // Qof (100)
	'r':  'ר', // Resh (200)
	't':  'ת', // Tav (400)   — plain dental; Tet has no ASCII equivalent
}

// academicVowels is the set of single-byte vowel characters that are silently
// dropped in the academic scheme (no consonantal representation).
var academicVowels = map[byte]bool{
	'a': true,
	'e': true,
	'i': true,
	'o': true,
	'u': true,
}

// sofitMap maps each base letter that has a sofit (final) form to that form.
// Both schemes apply the same sofit substitution rule.
var sofitMap = map[rune]rune{
	'כ': 'ך', // Kaf → Kaf Sofit
	'מ': 'ם', // Mem → Mem Sofit
	'נ': 'ן', // Nun → Nun Sofit
	'פ': 'ף', // Pe  → Pe Sofit
	'צ': 'ץ', // Tsade → Tsade Sofit
}

// transliterateAcademic converts a single word-part (no spaces) from Latin
// to a slice of Hebrew runes using the academic (strict consonantal) scheme.
//
// Algorithm:
//  1. Lowercase the input.
//  2. Greedy left-to-right scan: try 2-char multi-char sequence first, then
//     single char. Vowels (a/e/i/o/u) are dropped silently.
//  3. An unrecognised character (not a consonant, not a vowel, not in the
//     multi-char table) causes an immediate error.
//  4. After all characters are processed, apply sofit substitution: if the
//     last rune in the output is one of {כ,מ,נ,פ,צ}, replace it with its
//     sofit form.
//  5. If the resulting rune slice is empty (e.g., the input was all vowels),
//     return an error.
func transliterateAcademic(input string) ([]rune, error) {
	s := strings.ToLower(input)
	var out []rune

	for i := 0; i < len(s); {
		// Try 2-char multi-char sequence first.
		if i+1 < len(s) {
			pair := s[i : i+2]
			if r, ok := academicMultiChar[pair]; ok {
				out = append(out, r)
				i += 2
				continue
			}
		}
		// Try single char.
		c := s[i]
		if academicVowels[c] {
			// Silently drop vowels.
			i++
			continue
		}
		if r, ok := academicSingleChar[c]; ok {
			out = append(out, r)
			i++
			continue
		}
		// Unmappable character.
		return nil, fmt.Errorf("academic scheme: cannot map character %q in input %q", string(c), input)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("academic scheme: input %q produces no Hebrew letters (all vowels or empty)", input)
	}

	// Apply sofit substitution to the last letter.
	if sofit, ok := sofitMap[out[len(out)-1]]; ok {
		out[len(out)-1] = sofit
	}

	return out, nil
}
