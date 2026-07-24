package gematria

import (
	"errors"
	"math"
	"sort"
	"strings"
)

// DefaultLookupLimit is the default maximum number of results returned
// by FindByValue when limit is zero or negative.
const DefaultLookupLimit = 20

// FindByValue finds words in source whose gematria value under system
// equals value, returning at most limit results.
//
// If limit is zero or negative, DefaultLookupLimit (20) is used.
// Returns (nil, false, non-nil error) if source is nil.
// Otherwise returns the result of source.FindByValue directly.
func FindByValue(value int, source WordSource, system System, limit int) ([]Word, bool, error) {
	if source == nil {
		return nil, false, errors.New("gematria: FindByValue called with nil WordSource")
	}
	if limit <= 0 {
		limit = DefaultLookupLimit
	}
	return source.FindByValue(value, system, limit)
}

// aliases maps lowercase transliteration strings to their Hebrew rune.
// Populated once by init() from the letters dictionary.
var aliases map[string]rune

func init() {
	aliases = make(map[string]rune, 60)
	for r, l := range letters {
		for _, alias := range l.Aliases {
			aliases[alias] = r
		}
	}
}

// LookupLetter returns the Letter for the given Hebrew rune.
// Returns *InvalidCharError if r is not a recognized Hebrew character.
//
// Note: the function is named LookupLetter rather than Letter to avoid a
// package-level identifier conflict with the Letter struct type.
func LookupLetter(r rune) (Letter, error) {
	l, ok := letters[r]
	if !ok {
		return Letter{}, &InvalidCharError{Char: r}
	}
	return l, nil
}

// LetterByName returns the Letter for the given transliteration alias.
// The lookup is case-insensitive. Returns *UnknownNameError (with Suggestions)
// if no alias matches.
func LetterByName(name string) (Letter, error) {
	key := strings.ToLower(name)
	if r, ok := aliases[key]; ok {
		return letters[r], nil
	}
	suggestions := suggestAliases(key)
	return Letter{}, &UnknownNameError{Name: name, Suggestions: suggestions}
}

// AtbashSubstitute returns the Atbash mirror rune for r.
// If r is not a recognized Hebrew letter, r is returned unchanged.
func AtbashSubstitute(r rune) rune {
	if mirror, ok := atbashMirror[r]; ok {
		return mirror
	}
	return r
}

// suggestAliases returns alias strings whose Levenshtein distance from input
// is > 0 and <= min(ceil(len(input)/2), 2). Results are sorted for stability.
func suggestAliases(input string) []string {
	maxDist := min(int(math.Ceil(float64(len(input))/2)), 2)
	seen := map[string]bool{}
	var out []string
	for alias := range aliases {
		d := levenshtein(input, alias)
		if d > 0 && d <= maxDist && !seen[alias] {
			out = append(out, alias)
			seen[alias] = true
		}
	}
	sort.Strings(out)
	return out
}

// levenshtein computes the edit distance between a and b operating on runes
// so multi-byte Unicode characters are counted as single units.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	m, n := len(ra), len(rb)
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			if ra[i-1] == rb[j-1] {
				curr[j] = prev[j-1]
			} else {
				curr[j] = 1 + minInt(prev[j], curr[j-1], prev[j-1])
			}
		}
		prev, curr = curr, prev
	}
	return prev[n]
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// isHebrewRune reports whether r is in the Hebrew letter Unicode block
// (U+05D0–U+05EA), which covers all 22 standard letters and 5 sofit forms.
func isHebrewRune(r rune) bool {
	return r >= '\u05D0' && r <= '\u05EA'
}

// Parse resolves an input string into a sequence of Hebrew letters.
// It handles three input modes automatically:
//   - Pure Hebrew Unicode: iterates rune-by-rune; returns *InvalidCharError for
//     any rune outside the Hebrew block.
//   - Pure Latin transliteration: space-splits the string; returns
//     *UnknownNameError (with Levenshtein suggestions) on no alias match.
//   - Mixed: Hebrew runes are treated as single-rune tokens; Latin words are
//     space-separated and looked up by alias.
//
// Position in errors: byte offset for Hebrew-input errors; 0-based token index
// for Latin-input errors.
func Parse(input string) ([]Letter, error) {
	if input == "" {
		return nil, nil
	}

	// Pre-scan to determine input mode.
	var hasHebrew, hasLatin bool
	for _, r := range input {
		if r == ' ' {
			continue
		}
		if isHebrewRune(r) {
			hasHebrew = true
		} else {
			hasLatin = true
		}
	}

	switch {
	case !hasHebrew && !hasLatin:
		return nil, nil
	case hasHebrew && !hasLatin:
		return parsePureHebrew(input)
	case !hasHebrew && hasLatin:
		return parsePureLatin(input)
	default:
		return parseMixed(input)
	}
}

// parsePureHebrew processes a string that contains only Hebrew runes (and
// optional spaces). Any non-Hebrew, non-space rune is reported as an
// *InvalidCharError with its byte offset.
func parsePureHebrew(input string) ([]Letter, error) {
	var out []Letter
	for byteOffset, r := range input {
		if r == ' ' {
			continue
		}
		l, err := LookupLetter(r)
		if err != nil {
			return nil, &InvalidCharError{Char: r, Position: byteOffset, Input: input}
		}
		out = append(out, l)
	}
	return out, nil
}

// parsePureLatin processes a string that contains only Latin characters.
// It splits by spaces and calls LetterByName on each token, setting the
// token index (0-based) as Position on any *UnknownNameError.
func parsePureLatin(input string) ([]Letter, error) {
	tokens := strings.Fields(input)
	out := make([]Letter, 0, len(tokens))
	for idx, tok := range tokens {
		l, err := LetterByName(tok)
		if err != nil {
			var une *UnknownNameError
			if errors.As(err, &une) {
				une.Position = idx
			}
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// parseMixed processes a string that contains both Hebrew runes and Latin
// characters. Hebrew runes are single-rune tokens; Latin characters
// accumulate into space-separated tokens.
func parseMixed(input string) ([]Letter, error) {
	var out []Letter
	var latinBuf strings.Builder
	latinIdx := 0

	flushLatin := func() error {
		tok := strings.TrimSpace(latinBuf.String())
		latinBuf.Reset()
		if tok == "" {
			return nil
		}
		l, err := LetterByName(tok)
		if err != nil {
			var une *UnknownNameError
			if errors.As(err, &une) {
				une.Position = latinIdx
			}
			return err
		}
		out = append(out, l)
		latinIdx++
		return nil
	}

	for _, r := range input {
		if isHebrewRune(r) {
			if err := flushLatin(); err != nil {
				return nil, err
			}
			l, _ := LookupLetter(r) // always succeeds for runes in Hebrew range
			out = append(out, l)
		} else if r == ' ' {
			if latinBuf.Len() > 0 {
				if err := flushLatin(); err != nil {
					return nil, err
				}
			}
		} else {
			latinBuf.WriteRune(r)
		}
	}
	if err := flushLatin(); err != nil {
		return nil, err
	}
	return out, nil
}

// validateSystem returns *InvalidSystemError if s is not a recognized system.
func validateSystem(s System) error {
	if _, ok := systemValues[s]; ok {
		return nil
	}
	return &InvalidSystemError{Name: string(s), Valid: ValidSystems()}
}

// ComputeFromLetters computes a Result from a pre-resolved letter sequence
// under the given system.
//
// It is the primitive used by both Compute (which calls Parse first) and
// ComputeTransliterated (which calls Transliterate first). Library consumers
// can call it directly when they want to inspect or modify the letter
// sequence between resolution and computation.
//
// For the Atbash system, the value tables already map each original rune to
// its mirror's Hechrachi value, so no substitution is performed at compute
// time. AtbashSubstitute is provided separately for display purposes.
//
// The returned Result has Scheme set to the zero value (empty Scheme).
// Callers that performed transliteration should set it on the returned Result
// — see ComputeTransliterated.
//
// Returns *InvalidSystemError if system is not recognized.
func ComputeFromLetters(input string, ltrs []Letter, system System) (Result, error) {
	if err := validateSystem(system); err != nil {
		return Result{}, err
	}
	table := systemValues[system]
	letterResults := make([]LetterResult, len(ltrs))
	total := 0
	for i, l := range ltrs {
		val := table[l.Char]
		letterResults[i] = LetterResult{Letter: l, Value: val}
		total += val
	}
	return Result{
		Input:   input,
		System:  system,
		Total:   total,
		Letters: letterResults,
	}, nil
}

// Compute computes the gematria value of input using the given system.
// It calls Parse to resolve the input into letters, then ComputeFromLetters
// to apply the system value table.
//
// Returns *InvalidSystemError if system is not recognized; or any error
// returned by Parse (typically *InvalidCharError or *UnknownNameError).
func Compute(input string, system System) (Result, error) {
	ltrs, err := Parse(input)
	if err != nil {
		return Result{}, err
	}
	return ComputeFromLetters(input, ltrs, system)
}

// ComputeTransliterated transliterates input using the given scheme, then
// computes its gematria value using the given system. Convenience wrapper
// around Transliterate followed by ComputeFromLetters; the returned Result
// has Scheme set to the supplied scheme.
//
// Returns *InvalidSchemeError or *UnknownWordError from the transliteration
// step, or *InvalidSystemError from the compute step.
func ComputeTransliterated(input string, system System, scheme Scheme) (Result, error) {
	ltrs, err := Transliterate(input, scheme)
	if err != nil {
		return Result{}, err
	}
	r, err := ComputeFromLetters(input, ltrs, system)
	if err != nil {
		return Result{}, err
	}
	r.Scheme = scheme
	return r, nil
}
