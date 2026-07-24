package cli

import (
	"fmt"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

func TestNewFormatter_returnsNonNil(t *testing.T) {
	formats := []string{"line", "value", "card", "json"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			f := NewFormatter(format, false, false)
			if f == nil {
				t.Errorf("NewFormatter(%q, false, false) returned nil", format)
			}
		})
	}
}

func TestNewFormatter_colorVariants(t *testing.T) {
	// Both color=true and color=false should return non-nil formatters.
	for _, useColor := range []bool{false, true} {
		t.Run(fmt.Sprintf("useColor=%v", useColor), func(t *testing.T) {
			for _, format := range []string{"line", "value", "card", "json"} {
				f := NewFormatter(format, useColor, false)
				if f == nil {
					t.Errorf("NewFormatter(%q, %v, false) returned nil", format, useColor)
				}
			}
		})
	}
}

func TestNewFormatter_defaultToLine(t *testing.T) {
	// Unknown format string falls back to line formatter (never panics).
	f := NewFormatter("unknown", false, false)
	if f == nil {
		t.Error("NewFormatter(unknown) returned nil, want lineFormatter fallback")
	}
}

// --- ValueFormatter ---

func makeResult(input string, system gematria.System) gematria.Result {
	r, err := gematria.Compute(input, system)
	if err != nil {
		panic("makeResult: " + err.Error())
	}
	return r
}

func TestValueFormatter_FormatResult_singleLetter(t *testing.T) {
	f := NewFormatter("value", false, false)
	r := makeResult("א", gematria.Hechrachi) // aleph = 1
	got := f.FormatResult(r)
	if got != "1\n" {
		t.Errorf("FormatResult(aleph) = %q, want %q", got, "1\n")
	}
}

func TestValueFormatter_FormatResult_word(t *testing.T) {
	f := NewFormatter("value", false, false)
	r := makeResult("שלום", gematria.Hechrachi) // shin(300)+lamed(30)+vav(6)+mem-sofit(40) = 376
	got := f.FormatResult(r)
	if got != "376\n" {
		t.Errorf("FormatResult(shalom) = %q, want %q", got, "376\n")
	}
}

func TestValueFormatter_FormatError(t *testing.T) {
	f := NewFormatter("value", false, false)
	err := &gematria.InvalidCharError{Char: 'x', Position: 0, Input: "x"}
	got := f.FormatError(err)
	if !strings.HasPrefix(got, "Error: ") {
		t.Errorf("FormatError = %q, want prefix %q", got, "Error: ")
	}
}

// --- LineFormatter ---

func TestLineFormatter_FormatResult_singleLetter(t *testing.T) {
	f := NewFormatter("line", false, false)
	r := makeResult("א", gematria.Hechrachi) // aleph = 1
	got := f.FormatResult(r)
	// Expected: "Aleph (‏א‎) = 1\n"
	want := "Aleph (" + rtlMark + "א" + ltrMark + ") = 1\n"
	if got != want {
		t.Errorf("FormatResult(aleph):\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLineFormatter_FormatResult_word(t *testing.T) {
	f := NewFormatter("line", false, false)
	r := makeResult("שלום", gematria.Hechrachi) // 376
	got := f.FormatResult(r)
	// Must contain the Hebrew word wrapped in RTL/LTR marks
	if !strings.Contains(got, rtlMark+"שלום"+ltrMark) {
		t.Errorf("FormatResult(shalom): missing RTL-wrapped word in %q", got)
	}
	// Must contain total
	if !strings.Contains(got, "376") {
		t.Errorf("FormatResult(shalom): missing total 376 in %q", got)
	}
	// Must contain breakdown: each letter RTL-wrapped with its value
	if !strings.Contains(got, rtlMark+"ש"+ltrMark+"=300") {
		t.Errorf("FormatResult(shalom): missing shin breakdown in %q", got)
	}
}

func TestLineFormatter_FormatResult_atbash(t *testing.T) {
	f := NewFormatter("line", false, true) // showAtbash=true
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	// Should append "→ ‏ת‎" (atbash of aleph is tav)
	if !strings.Contains(got, "\u2192") {
		t.Errorf("FormatResult with atbash: missing → in %q", got)
	}
	if !strings.Contains(got, rtlMark+"ת"+ltrMark) {
		t.Errorf("FormatResult with atbash: missing RTL-wrapped tav in %q", got)
	}
}

func TestLineFormatter_FormatResult_wordAtbash(t *testing.T) {
	f := NewFormatter("line", false, true)
	r := makeResult("אב", gematria.Hechrachi) // aleph + bet
	got := f.FormatResult(r)
	// atbash of aleph=tav, bet=shin → "תש"
	if !strings.Contains(got, "\u2192") {
		t.Errorf("FormatResult word atbash: missing → in %q", got)
	}
	// The mirrors concatenated should appear somewhere
	if !strings.Contains(got, "ת") || !strings.Contains(got, "ש") {
		t.Errorf("FormatResult word atbash: missing mirror letters in %q", got)
	}
}

func TestLineFormatter_FormatError_plain(t *testing.T) {
	f := NewFormatter("line", false, false)
	err := &gematria.InvalidCharError{Char: 'x', Position: 0, Input: "x"}
	got := f.FormatError(err)
	if !strings.HasPrefix(got, "Error: ") {
		t.Errorf("FormatError = %q, want prefix %q", got, "Error: ")
	}
}

// --- CardFormatter ---

func TestCardFormatter_FormatResult_containsHeader(t *testing.T) {
	f := NewFormatter("card", false, false)
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	if !strings.Contains(got, "Letter") {
		t.Errorf("card FormatResult: missing 'Letter' header in:\n%s", got)
	}
	if !strings.Contains(got, "Name") {
		t.Errorf("card FormatResult: missing 'Name' header in:\n%s", got)
	}
	if !strings.Contains(got, "Value") {
		t.Errorf("card FormatResult: missing 'Value' header in:\n%s", got)
	}
	if !strings.Contains(got, "Meaning") {
		t.Errorf("card FormatResult: missing 'Meaning' header in:\n%s", got)
	}
}

func TestCardFormatter_FormatResult_containsLetterData(t *testing.T) {
	f := NewFormatter("card", false, false)
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	if !strings.Contains(got, "Aleph") {
		t.Errorf("card FormatResult: missing letter name 'Aleph' in:\n%s", got)
	}
	if !strings.Contains(got, "1") {
		t.Errorf("card FormatResult: missing value '1' in:\n%s", got)
	}
	if !strings.Contains(got, "ox") {
		t.Errorf("card FormatResult: missing meaning 'ox' in:\n%s", got)
	}
}

func TestCardFormatter_FormatResult_containsTotal(t *testing.T) {
	f := NewFormatter("card", false, false)
	r := makeResult("שלום", gematria.Hechrachi)
	got := f.FormatResult(r)
	if !strings.Contains(got, "Total:") {
		t.Errorf("card FormatResult: missing 'Total:' in:\n%s", got)
	}
	if !strings.Contains(got, "376") {
		t.Errorf("card FormatResult: missing total 376 in:\n%s", got)
	}
	if !strings.Contains(got, "hechrachi") {
		t.Errorf("card FormatResult: missing system name in:\n%s", got)
	}
}

func TestCardFormatter_FormatResult_containsSeparator(t *testing.T) {
	f := NewFormatter("card", false, false)
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	if !strings.Contains(got, "\u2500") {
		t.Errorf("card FormatResult: missing separator '─' in:\n%s", got)
	}
}

func TestCardFormatter_FormatResult_atbashColumn(t *testing.T) {
	f := NewFormatter("card", false, true) // showAtbash=true
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	if !strings.Contains(got, "Atbash") {
		t.Errorf("card FormatResult with atbash: missing 'Atbash' column header in:\n%s", got)
	}
	// aleph → tav in atbash
	if !strings.Contains(got, "ת") {
		t.Errorf("card FormatResult with atbash: missing tav mirror in:\n%s", got)
	}
}

func TestCardFormatter_FormatResult_hebrewRTL(t *testing.T) {
	f := NewFormatter("card", false, false)
	r := makeResult("ב", gematria.Hechrachi)
	got := f.FormatResult(r)
	// Hebrew char in rows must be RTL-wrapped
	if !strings.Contains(got, rtlMark+"ב"+ltrMark) {
		t.Errorf("card FormatResult: Hebrew char not RTL-wrapped in:\n%s", got)
	}
}

// --- JSONFormatter ---

func TestJSONFormatter_FormatResult_singleLetter(t *testing.T) {
	f := NewFormatter("json", false, false)
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)

	// Must be valid JSON ending with newline
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("JSON FormatResult: missing trailing newline in %q", got)
	}
	trimmed := strings.TrimSuffix(got, "\n")
	// Must start and end with braces (JSON object)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		t.Errorf("JSON FormatResult: not a JSON object: %q", trimmed)
	}
	// Must contain required fields
	for _, field := range []string{`"input"`, `"system"`, `"total"`, `"letters"`} {
		if !strings.Contains(got, field) {
			t.Errorf("JSON FormatResult: missing field %s in %q", field, got)
		}
	}
	// Must contain correct values
	if !strings.Contains(got, `"input":"א"`) {
		t.Errorf("JSON FormatResult: wrong input field in %q", got)
	}
	if !strings.Contains(got, `"system":"hechrachi"`) {
		t.Errorf("JSON FormatResult: wrong system field in %q", got)
	}
	if !strings.Contains(got, `"total":1`) {
		t.Errorf("JSON FormatResult: wrong total in %q", got)
	}
}

func TestJSONFormatter_FormatResult_letterFields(t *testing.T) {
	f := NewFormatter("json", false, false)
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)

	// Per-letter fields
	for _, field := range []string{`"char"`, `"name"`, `"value"`, `"meaning"`, `"position"`} {
		if !strings.Contains(got, field) {
			t.Errorf("JSON FormatResult: missing letter field %s in %q", field, got)
		}
	}
	if !strings.Contains(got, `"char":"א"`) {
		t.Errorf("JSON FormatResult: wrong char in %q", got)
	}
	if !strings.Contains(got, `"name":"Aleph"`) {
		t.Errorf("JSON FormatResult: wrong name in %q", got)
	}
	if !strings.Contains(got, `"position":1`) {
		t.Errorf("JSON FormatResult: wrong position in %q", got)
	}
}

func TestJSONFormatter_FormatResult_noRTLMarks(t *testing.T) {
	f := NewFormatter("json", false, false)
	r := makeResult("שלום", gematria.Hechrachi)
	got := f.FormatResult(r)
	if strings.Contains(got, rtlMark) || strings.Contains(got, ltrMark) {
		t.Errorf("JSON FormatResult: must not contain RTL/LTR marks: %q", got)
	}
}

func TestJSONFormatter_FormatResult_noANSI(t *testing.T) {
	f := NewFormatter("json", true, true) // color=true should have no effect
	r := makeResult("א", gematria.Hechrachi)
	got := f.FormatResult(r)
	if strings.Contains(got, "\033[") {
		t.Errorf("JSON FormatResult: must not contain ANSI codes: %q", got)
	}
}

func TestJSONFormatter_FormatError_invalidChar(t *testing.T) {
	f := NewFormatter("json", false, false)
	err := &gematria.InvalidCharError{Char: 'x', Position: 3, Input: "אאx"}
	got := f.FormatError(err)

	if !strings.Contains(got, `"error"`) {
		t.Errorf("JSON FormatError: missing 'error' field in %q", got)
	}
	if !strings.Contains(got, `"invalid_input":"x"`) {
		t.Errorf("JSON FormatError: wrong invalid_input in %q", got)
	}
	if !strings.Contains(got, `"position":3`) {
		t.Errorf("JSON FormatError: wrong position in %q", got)
	}
	if !strings.Contains(got, `"suggestions":[]`) {
		t.Errorf("JSON FormatError: missing empty suggestions array in %q", got)
	}
}

func TestJSONFormatter_FormatError_unknownName(t *testing.T) {
	f := NewFormatter("json", false, false)
	err := &gematria.UnknownNameError{Name: "shen", Position: 0, Suggestions: []string{"shin"}}
	got := f.FormatError(err)

	if !strings.Contains(got, `"invalid_input":"shen"`) {
		t.Errorf("JSON FormatError: wrong invalid_input in %q", got)
	}
	if !strings.Contains(got, `"suggestions":["shin"]`) {
		t.Errorf("JSON FormatError: wrong suggestions in %q", got)
	}
}

func TestJSONFormatter_FormatError_noSuggestions(t *testing.T) {
	f := NewFormatter("json", false, false)
	err := &gematria.UnknownNameError{Name: "xyzzy", Position: 0, Suggestions: nil}
	got := f.FormatError(err)

	// nil suggestions must serialize as [] not null
	if !strings.Contains(got, `"suggestions":[]`) {
		t.Errorf("JSON FormatError: nil suggestions must be [] not null, got %q", got)
	}
}

// --- FormatLookup ---

func TestLineFormatter_FormatLookup_basic(t *testing.T) {
	f := NewFormatter("line", false, false)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, false)
	want := rtlMark + "שלום" + ltrMark + "\n"
	if got != want {
		t.Errorf("FormatLookup basic = %q, want %q", got, want)
	}
}

func TestLineFormatter_FormatLookup_withTransliterationAndMeaning(t *testing.T) {
	f := NewFormatter("line", false, false)
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	got := f.FormatLookup(words, false)
	want := rtlMark + "שלום" + ltrMark + " (shalom) \u2014 peace\n"
	if got != want {
		t.Errorf("FormatLookup with meta = %q, want %q", got, want)
	}
}

func TestLineFormatter_FormatLookup_noResults(t *testing.T) {
	f := NewFormatter("line", false, false)
	got := f.FormatLookup(nil, false)
	if got != "(no results)\n" {
		t.Errorf("FormatLookup(nil) = %q, want %q", got, "(no results)\n")
	}
}

func TestLineFormatter_FormatLookup_hasMore(t *testing.T) {
	f := NewFormatter("line", false, false)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, true)
	if !strings.HasSuffix(got, "(more results available \u2014 increase --limit to see them)\n") {
		t.Errorf("FormatLookup hasMore: missing indicator in %q", got)
	}
}

func TestCardFormatter_FormatLookup_header(t *testing.T) {
	f := NewFormatterWithLookup("card", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"}}
	got := f.FormatLookup(words, false)
	if !strings.Contains(got, "Reverse Lookup: 376") {
		t.Errorf("card FormatLookup: missing header in:\n%s", got)
	}
	if !strings.Contains(got, "System: hechrachi") {
		t.Errorf("card FormatLookup: missing system in:\n%s", got)
	}
}

func TestCardFormatter_FormatLookup_numberedEntry(t *testing.T) {
	f := NewFormatterWithLookup("card", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{
		{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"},
		{Hebrew: "אמת"},
	}
	got := f.FormatLookup(words, false)
	if !strings.Contains(got, "1.") {
		t.Errorf("card FormatLookup: missing entry number 1 in:\n%s", got)
	}
	if !strings.Contains(got, "2.") {
		t.Errorf("card FormatLookup: missing entry number 2 in:\n%s", got)
	}
	if !strings.Contains(got, "Transliteration: shalom") {
		t.Errorf("card FormatLookup: missing transliteration in:\n%s", got)
	}
	if !strings.Contains(got, "Meaning: peace") {
		t.Errorf("card FormatLookup: missing meaning in:\n%s", got)
	}
	// Entry without transliteration/meaning should not show those lines
	if strings.Contains(got, "Transliteration: \n") || strings.Contains(got, "Meaning: \n") {
		t.Errorf("card FormatLookup: empty fields should not be shown:\n%s", got)
	}
}

func TestCardFormatter_FormatLookup_noResults(t *testing.T) {
	f := NewFormatterWithLookup("card", false, false, 376, gematria.Hechrachi)
	got := f.FormatLookup(nil, false)
	if got != "(no results)\n" {
		t.Errorf("card FormatLookup(nil) = %q, want %q", got, "(no results)\n")
	}
}

func TestCardFormatter_FormatLookup_hasMore(t *testing.T) {
	f := NewFormatterWithLookup("card", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, true)
	if !strings.HasSuffix(got, "(more results available \u2014 increase --limit to see them)\n") {
		t.Errorf("card FormatLookup hasMore: missing indicator in:\n%s", got)
	}
}

func TestJSONFormatter_FormatLookup_structure(t *testing.T) {
	f := NewFormatterWithLookup("json", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{
		{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"},
	}
	got := f.FormatLookup(words, false)
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("JSON FormatLookup: missing trailing newline in %q", got)
	}
	if !strings.Contains(got, `"value":376`) {
		t.Errorf("JSON FormatLookup: missing value field in %q", got)
	}
	if !strings.Contains(got, `"system":"hechrachi"`) {
		t.Errorf("JSON FormatLookup: missing system field in %q", got)
	}
	if !strings.Contains(got, `"results"`) {
		t.Errorf("JSON FormatLookup: missing results field in %q", got)
	}
	if !strings.Contains(got, `"hasMore":false`) {
		t.Errorf("JSON FormatLookup: missing hasMore=false in %q", got)
	}
	if !strings.Contains(got, `"word":"שלום"`) {
		t.Errorf("JSON FormatLookup: missing word field in %q", got)
	}
	if !strings.Contains(got, `"transliteration":"shalom"`) {
		t.Errorf("JSON FormatLookup: missing transliteration in %q", got)
	}
	if !strings.Contains(got, `"meaning":"peace"`) {
		t.Errorf("JSON FormatLookup: missing meaning in %q", got)
	}
}

func TestJSONFormatter_FormatLookup_emptyResults(t *testing.T) {
	f := NewFormatterWithLookup("json", false, false, 376, gematria.Hechrachi)
	got := f.FormatLookup(nil, false)
	if !strings.Contains(got, `"results":[]`) {
		t.Errorf("JSON FormatLookup empty: results must be [] not null, got %q", got)
	}
	if !strings.Contains(got, `"hasMore":false`) {
		t.Errorf("JSON FormatLookup empty: missing hasMore in %q", got)
	}
}

func TestJSONFormatter_FormatLookup_hasMore(t *testing.T) {
	f := NewFormatterWithLookup("json", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, true)
	if !strings.Contains(got, `"hasMore":true`) {
		t.Errorf("JSON FormatLookup hasMore: got %q", got)
	}
}

func TestJSONFormatter_FormatLookup_noRTLMarks(t *testing.T) {
	f := NewFormatterWithLookup("json", false, false, 376, gematria.Hechrachi)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, false)
	if strings.Contains(got, rtlMark) || strings.Contains(got, ltrMark) {
		t.Errorf("JSON FormatLookup: must not contain RTL/LTR marks in %q", got)
	}
}

func TestJSONFormatter_FormatLookup_noANSI(t *testing.T) {
	f := NewFormatterWithLookup("json", true, true, 376, gematria.Hechrachi)
	words := []gematria.Word{{Hebrew: "שלום"}}
	got := f.FormatLookup(words, false)
	if strings.Contains(got, "\033[") {
		t.Errorf("JSON FormatLookup: must not contain ANSI codes in %q", got)
	}
}

func TestValueFormatter_FormatLookup_empty(t *testing.T) {
	f := NewFormatter("value", false, false)
	got := f.FormatLookup(nil, false)
	if got != "" {
		t.Errorf("FormatLookup(nil): got %q, want empty string", got)
	}
}

func TestValueFormatter_FormatLookup_words(t *testing.T) {
	f := NewFormatter("value", false, false)
	words := []gematria.Word{
		{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"},
		{Hebrew: "אמת", Transliteration: "emet", Meaning: "truth"},
	}
	got := f.FormatLookup(words, false)
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("FormatLookup: got %d lines, want 2:\n%q", len(lines), got)
	}
	if lines[0] != "שלום" {
		t.Errorf("line 0 = %q, want %q", lines[0], "שלום")
	}
	if lines[1] != "אמת" {
		t.Errorf("line 1 = %q, want %q", lines[1], "אמת")
	}
}
