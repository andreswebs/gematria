package gematria

import (
	"errors"
	"testing"
)

// transliterateAcademic tests

// academicRunes is a helper for academic table-driven tests.
func academicRunes(t *testing.T, input string, want []rune) {
	t.Helper()
	got, err := transliterateAcademic(input)
	if err != nil {
		t.Fatalf("transliterateAcademic(%q) unexpected error: %v", input, err)
	}
	if string(got) != string(want) {
		t.Errorf("transliterateAcademic(%q) = %q, want %q", input, string(got), string(want))
	}
}

func TestTransliterateAcademicMultiCharSh(t *testing.T) {
	academicRunes(t, "sh", []rune{'ש'})
}

func TestTransliterateAcademicMultiCharSequences(t *testing.T) {
	cases := []struct {
		input string
		want  []rune
	}{
		{"kh", []rune{'ח'}},
		{"ch", []rune{'ח'}},
		{"ts", []rune{'צ'}}, // sofit: צ→ץ
		{"tz", []rune{'צ'}}, // sofit: צ→ץ
		{"ph", []rune{'פ'}}, // sofit: פ→ף
	}
	// Note: single-rune words get sofit applied. ts/tz/ph end with sofit forms.
	sofitExpect := map[string]rune{
		"ts": 'ץ', "tz": 'ץ', "ph": 'ף',
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := transliterateAcademic(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("expected 1 rune, got %d: %q", len(got), string(got))
			}
			if expected, hasSofit := sofitExpect[tc.input]; hasSofit {
				if got[0] != expected {
					t.Errorf("transliterateAcademic(%q) = %q, want %q", tc.input, string(got), string([]rune{expected}))
				}
			} else {
				if got[0] != tc.want[0] {
					t.Errorf("transliterateAcademic(%q) = %q, want %q", tc.input, string(got), string(tc.want))
				}
			}
		})
	}
}

func TestTransliterateAcademicVowelsDropped(t *testing.T) {
	// Single vowels should produce empty sequence → error.
	for _, vowel := range []string{"a", "e", "i", "o", "u"} {
		t.Run(vowel, func(t *testing.T) {
			_, err := transliterateAcademic(vowel)
			if err == nil {
				t.Errorf("transliterateAcademic(%q) expected error for all-vowel input", vowel)
			}
		})
	}
}

func TestTransliterateAcademicSofitTransformation(t *testing.T) {
	cases := []struct {
		input string
		want  rune // last rune should be sofit
	}{
		{"shalom", 'ם'}, // מ→ם: sh-a(drop)-l-o(drop)-m
		{"nun", 'ן'},    // נ→ן: n-u(drop)-n
		{"kaf", 'ף'},    // No wait - "kaf" = k-a(drop)-f = כפ then sofit: פ→ף
		{"yats", 'ץ'},   // y-a(drop)-ts = יצ then sofit: צ→ץ
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := transliterateAcademic(tc.input)
			if err != nil {
				t.Fatalf("transliterateAcademic(%q) unexpected error: %v", tc.input, err)
			}
			last := got[len(got)-1]
			if last != tc.want {
				t.Errorf("transliterateAcademic(%q) last rune = %q, want %q", tc.input, string(last), string(tc.want))
			}
		})
	}
}

func TestTransliterateAcademicFullWordExamples(t *testing.T) {
	// Examples from spec §4.1.4 and §4.5 (sofit applied).
	cases := []struct {
		input string
		want  string // Hebrew string (UTF-8)
	}{
		{"shalom", "שלם"},    // sh→ש a→drop l→ל o→drop m→מ→ם(sofit)
		{"gadol", "גדל"},     // g→ג a→drop d→ד o→drop l→ל (no sofit for ל)
		{"emet", "מת"},       // e→drop m→מ e→drop t→ת (no sofit for ת)
		{"bereshit", "ברשת"}, // b→ב e→drop r→ר e→drop sh→ש i→drop t→ת
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			academicRunes(t, tc.input, []rune(tc.want))
		})
	}
}

func TestTransliterateAcademicExplicitAleph(t *testing.T) {
	// '→א, then m→מ, e→drop, t→ת; no sofit for ת
	academicRunes(t, "'emet", []rune("אמת"))
}

func TestTransliterateAcademicASCIIFallbackX(t *testing.T) {
	// x is the ASCII fallback for ח (Het)
	academicRunes(t, "x", []rune{'ח'})
}

func TestTransliterateAcademicCaseInsensitive(t *testing.T) {
	lower, err1 := transliterateAcademic("shalom")
	upper, err2 := transliterateAcademic("SHALOM")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if string(lower) != string(upper) {
		t.Errorf("case insensitivity failed: lower=%q upper=%q", string(lower), string(upper))
	}
}

func TestTransliterateAcademicUnmappableReturnsError(t *testing.T) {
	cases := []string{"1", "0", "-", "_", "@", "!"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			_, err := transliterateAcademic(c)
			if err == nil {
				t.Errorf("transliterateAcademic(%q) expected error for unmappable input", c)
			}
		})
	}
}

func TestTransliterateAcademicAllVowelsError(t *testing.T) {
	_, err := transliterateAcademic("aeiou")
	if err == nil {
		t.Error("transliterateAcademic(\"aeiou\") expected error (all vowels → empty)")
	}
}

func TestTransliterateAcademicWAndVBothMapToVav(t *testing.T) {
	got1, _ := transliterateAcademic("v")
	got2, _ := transliterateAcademic("w")
	if string(got1) != string(got2) {
		t.Errorf("v and w should both map to Vav: v=%q w=%q", string(got1), string(got2))
	}
	if got1[0] != 'ו' {
		t.Errorf("v should map to Vav (ו), got %q", string(got1))
	}
}

// Scheme type and constants

func TestSchemeConstants(t *testing.T) {
	if SchemeAcademic != "academic" {
		t.Errorf("SchemeAcademic = %q, want %q", SchemeAcademic, "academic")
	}
	if SchemeIsraeli != "israeli" {
		t.Errorf("SchemeIsraeli = %q, want %q", SchemeIsraeli, "israeli")
	}
}

func TestValidSchemesReturnsBoth(t *testing.T) {
	schemes := ValidSchemes()
	if len(schemes) != 2 {
		t.Fatalf("ValidSchemes() returned %d schemes, want 2", len(schemes))
	}
	if schemes[0] != SchemeAcademic {
		t.Errorf("ValidSchemes()[0] = %q, want SchemeAcademic", schemes[0])
	}
	if schemes[1] != SchemeIsraeli {
		t.Errorf("ValidSchemes()[1] = %q, want SchemeIsraeli", schemes[1])
	}
}

func TestSchemeIsStringBased(t *testing.T) {
	// Scheme is a named string type — it must be directly comparable to a string constant.
	var s Scheme = "academic"
	if s != SchemeAcademic {
		t.Errorf("Scheme(%q) != SchemeAcademic", s)
	}
}

// ValidSchemes stable order

func TestValidSchemesStableOrder(t *testing.T) {
	first := ValidSchemes()
	second := ValidSchemes()
	if len(first) != len(second) {
		t.Fatalf("ValidSchemes() lengths differ: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("ValidSchemes()[%d] differs: %q vs %q", i, first[i], second[i])
		}
	}
}

// transliterateIsraeli tests

// israeliRunes is a helper for table-driven tests of the israeli scheme.
func israeliRunes(t *testing.T, input string, want string) {
	t.Helper()
	got, err := transliterateIsraeli(input)
	if err != nil {
		t.Fatalf("transliterateIsraeli(%q) unexpected error: %v", input, err)
	}
	if string(got) != want {
		t.Errorf("transliterateIsraeli(%q) = %q, want %q", input, string(got), want)
	}
}

// TestTransliterateIsraeliFullWordExamples covers the canonical examples from
// spec §4.2.4 and §4.3.
func TestTransliterateIsraeliFullWordExamples(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		// spec §4.2.4 examples
		{"shalom", "שלום"}, // sh→ש a(medial)→drop l→ל o(non-init)→ו m→מ→ם(sofit)
		{"gadol", "גדול"},  // g→ג a(medial)→drop d→ד o(non-init)→ו l→ל
		{"emet", "אמת"},    // e(initial)→א m→מ e(medial)→drop t→ת
		{"or", "אור"},      // o(initial)→או r→ר
		{"shira", "שירה"},  // sh→ש i→י r→ר a(final)→ה
		{"yafe", "יפה"},    // y→י a(medial)→drop f→פ e(final)→ה
		{"david", "דויד"},  // d→ד a(medial)→drop v→ו i→י d→ד
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			israeliRunes(t, tc.input, tc.want)
		})
	}
}

// TestTransliterateIsraeliVsAcademic verifies the scheme produces different
// results from academic when vowels are present (spec §4.3).
func TestTransliterateIsraeliVsAcademic(t *testing.T) {
	// shalom: academic drops o → שלם, israeli keeps it → שלום
	academic, err := transliterateAcademic("shalom")
	if err != nil {
		t.Fatalf("academic shalom error: %v", err)
	}
	israeli, err := transliterateIsraeli("shalom")
	if err != nil {
		t.Fatalf("israeli shalom error: %v", err)
	}
	if string(academic) == string(israeli) {
		t.Errorf("academic and israeli should differ for \"shalom\": both give %q", string(academic))
	}
	if string(israeli) != "שלום" {
		t.Errorf("israeli shalom = %q, want %q", string(israeli), "שלום")
	}
}

// TestTransliterateIsraeliAmbiguousCombos checks that ambiguous multi-char
// sequences resolve to the same letters as academic (spec §4.2.1 = same table).
func TestTransliterateIsraeliAmbiguousCombos(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"sh", "ש"}, // Shin — no sofit for ש
		{"ch", "ח"}, // Het via "ch"
		{"kh", "ח"}, // Het via "kh"
		{"ts", "ץ"}, // Tsade → sofit ץ (single-letter word)
		{"tz", "ץ"}, // Tsade alternate → sofit ץ
		{"ph", "ף"}, // Pe → sofit ף
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			israeliRunes(t, tc.input, tc.want)
		})
	}
}

// TestTransliterateIsraeliWordInitialVowels checks that word-initial a/e → א
// and word-initial o/u → או.
func TestTransliterateIsraeliWordInitialVowels(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"emet", "אמת"}, // e initial → א
		{"or", "אור"},   // o initial → או
		{"ul", "אול"},   // u initial → או then l
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			israeliRunes(t, tc.input, tc.want)
		})
	}
}

// TestTransliterateIsraeliWordFinalVowels checks that word-final a/e → ה.
func TestTransliterateIsraeliWordFinalVowels(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"shira", "שירה"}, // final a → ה
		{"yafe", "יפה"},   // final e → ה
		{"shana", "שנה"},  // sh→ש a(medial)→drop n→נ a(final)→ה
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			israeliRunes(t, tc.input, tc.want)
		})
	}
}

// TestTransliterateIsraeliYodMater checks that i always maps to Yod.
func TestTransliterateIsraeliYodMater(t *testing.T) {
	israeliRunes(t, "shira", "שירה") // i → י embedded in word
	israeliRunes(t, "ki", "כי")      // k→כ i→י (no sofit for י)
}

// TestTransliterateIsraeliSofitTransformation checks that sofit substitution
// is applied to the last letter of the output.
func TestTransliterateIsraeliSofitTransformation(t *testing.T) {
	cases := []struct {
		input string
		last  rune
	}{
		{"shalom", 'ם'}, // מ→ם
		{"amen", 'ן'},   // a(init)→א m→מ e(medial)→drop n→נ→ן
		{"yats", 'ץ'},   // y→י a(medial)→drop ts→צ→ץ
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := transliterateIsraeli(tc.input)
			if err != nil {
				t.Fatalf("transliterateIsraeli(%q) unexpected error: %v", tc.input, err)
			}
			last := got[len(got)-1]
			if last != tc.last {
				t.Errorf("transliterateIsraeli(%q) last rune = %q, want %q", tc.input, string(last), string(tc.last))
			}
		})
	}
}

// TestTransliterateIsraeliCaseInsensitive checks that input is case-folded.
func TestTransliterateIsraeliCaseInsensitive(t *testing.T) {
	lower, err1 := transliterateIsraeli("shalom")
	upper, err2 := transliterateIsraeli("SHALOM")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if string(lower) != string(upper) {
		t.Errorf("case insensitivity failed: lower=%q upper=%q", string(lower), string(upper))
	}
}

// TestTransliterateIsraeliUnmappableReturnsError checks that unknown chars
// cause an error.
func TestTransliterateIsraeliUnmappableReturnsError(t *testing.T) {
	cases := []string{"1", "0", "-", "_", "@", "!"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			_, err := transliterateIsraeli(c)
			if err == nil {
				t.Errorf("transliterateIsraeli(%q) expected error for unmappable input", c)
			}
		})
	}
}

// TestTransliterateIsraeliEmptyInputError checks that empty input returns an
// error.
func TestTransliterateIsraeliEmptyInputError(t *testing.T) {
	_, err := transliterateIsraeli("")
	if err == nil {
		t.Error("transliterateIsraeli(\"\") expected error for empty input")
	}
}

// --- Transliterate public function tests ---

// TestTransliterateAcademicDispatch is the tracer bullet: confirm that
// Transliterate routes to the academic scheme and returns the correct letters.
func TestTransliterateAcademicDispatch(t *testing.T) {
	// "shalom" academic → שלם (sh→ש, a→drop, l→ל, o→drop, m→מ→ם sofit)
	got, err := Transliterate("shalom", SchemeAcademic)
	if err != nil {
		t.Fatalf("Transliterate(%q, SchemeAcademic) unexpected error: %v", "shalom", err)
	}
	want := "שלם"
	var gotStr string
	for _, l := range got {
		gotStr += string(l.Char)
	}
	if gotStr != want {
		t.Errorf("Transliterate(%q, SchemeAcademic) = %q, want %q", "shalom", gotStr, want)
	}
}

// TestTransliteratePositionOnSecondTokenFailure verifies that when the second
// word-part fails, UnknownWordError.Position is 1 (0-based token index).
func TestTransliteratePositionOnSecondTokenFailure(t *testing.T) {
	// "shalom 1bad" — first token OK, second starts with digit (unmappable)
	_, err := Transliterate("shalom 1bad", SchemeAcademic)
	if err == nil {
		t.Fatal("expected error for second unmappable token, got nil")
	}
	var uwe *UnknownWordError
	if !errors.As(err, &uwe) {
		t.Fatalf("expected *UnknownWordError, got %T: %v", err, err)
	}
	if uwe.Position != 1 {
		t.Errorf("UnknownWordError.Position = %d, want 1 (second token)", uwe.Position)
	}
	if uwe.Input != "1bad" {
		t.Errorf("UnknownWordError.Input = %q, want %q", uwe.Input, "1bad")
	}
}

// TestTransliterateHebrewPassthrough verifies that a word-part consisting
// entirely of Hebrew Unicode characters is passed through via LookupLetter,
// regardless of scheme.
func TestTransliterateHebrewPassthrough(t *testing.T) {
	// שלום is direct Hebrew — should resolve to the same letters under both schemes.
	for _, scheme := range ValidSchemes() {
		t.Run(string(scheme), func(t *testing.T) {
			got, err := Transliterate("שלום", scheme)
			if err != nil {
				t.Fatalf("Transliterate(\"שלום\", %q) unexpected error: %v", scheme, err)
			}
			want := "שלום"
			var gotStr string
			for _, l := range got {
				gotStr += string(l.Char)
			}
			if gotStr != want {
				t.Errorf("Transliterate(\"שלום\", %q) = %q, want %q", scheme, gotStr, want)
			}
		})
	}
}

// TestTransliterateMultiToken verifies that a space-containing string is split
// into independent word-parts, each transliterated with sofit rules applied
// per-part, and the results concatenated.
func TestTransliterateMultiToken(t *testing.T) {
	// "shalom emet" academic:
	//   "shalom" → שלם (m→ם sofit)
	//   "emet"   → מת  (e→drop, m→מ, e→drop, t→ת — no sofit for ת)
	// Combined: שלםמת
	got, err := Transliterate("shalom emet", SchemeAcademic)
	if err != nil {
		t.Fatalf("Transliterate(\"shalom emet\", SchemeAcademic) unexpected error: %v", err)
	}
	want := "שלםמת"
	var gotStr string
	for _, l := range got {
		gotStr += string(l.Char)
	}
	if gotStr != want {
		t.Errorf("Transliterate(\"shalom emet\", SchemeAcademic) = %q, want %q", gotStr, want)
	}
	// Five letters total: ש ל ם מ ת
	if len(got) != 5 {
		t.Errorf("len(letters) = %d, want 5", len(got))
	}
}

// TestTransliterateIsraeliDispatch verifies that Transliterate routes to the
// israeli scheme, producing different results from academic for vowel-bearing input.
func TestTransliterateIsraeliDispatch(t *testing.T) {
	// "shalom" israeli → שלום (vowel 'o' kept as ו; academic drops it → שלם)
	got, err := Transliterate("shalom", SchemeIsraeli)
	if err != nil {
		t.Fatalf("Transliterate(%q, SchemeIsraeli) unexpected error: %v", "shalom", err)
	}
	want := "שלום"
	var gotStr string
	for _, l := range got {
		gotStr += string(l.Char)
	}
	if gotStr != want {
		t.Errorf("Transliterate(%q, SchemeIsraeli) = %q, want %q", "shalom", gotStr, want)
	}
}

// TestTransliterateUnmappableInput verifies that input containing characters
// that cannot be mapped returns *UnknownWordError with all fields populated.
func TestTransliterateUnmappableInput(t *testing.T) {
	_, err := Transliterate("sh@lom", SchemeAcademic)
	if err == nil {
		t.Fatal("Transliterate with unmappable char: expected error, got nil")
	}
	var uwe *UnknownWordError
	if !errors.As(err, &uwe) {
		t.Fatalf("expected *UnknownWordError, got %T: %v", err, err)
	}
	if uwe.Input != "sh@lom" {
		t.Errorf("UnknownWordError.Input = %q, want %q", uwe.Input, "sh@lom")
	}
	if uwe.Scheme != SchemeAcademic {
		t.Errorf("UnknownWordError.Scheme = %q, want SchemeAcademic", uwe.Scheme)
	}
	// Position 0: first (and only) token
	if uwe.Position != 0 {
		t.Errorf("UnknownWordError.Position = %d, want 0", uwe.Position)
	}
}

// TestTransliterateInvalidScheme verifies that an unrecognised scheme returns
// *InvalidSchemeError with the submitted name and the list of valid schemes.
func TestTransliterateInvalidScheme(t *testing.T) {
	_, err := Transliterate("shalom", Scheme("bogus"))
	if err == nil {
		t.Fatal("Transliterate with invalid scheme: expected error, got nil")
	}
	var ise *InvalidSchemeError
	if !errors.As(err, &ise) {
		t.Fatalf("expected *InvalidSchemeError, got %T: %v", err, err)
	}
	if ise.Name != "bogus" {
		t.Errorf("InvalidSchemeError.Name = %q, want %q", ise.Name, "bogus")
	}
	if len(ise.Valid) == 0 {
		t.Error("InvalidSchemeError.Valid is empty, want non-empty list")
	}
}
