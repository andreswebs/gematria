package gematria

import (
	"errors"
	"strings"
	"testing"
)

// --- ParseWordList: plain text ---

func TestParseWordList_PlainTextSingleWord(t *testing.T) {
	src, err := ParseWordList(strings.NewReader("שלום\n"))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, hasMore, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if hasMore {
		t.Error("hasMore = true, want false")
	}
	if len(words) != 1 {
		t.Fatalf("len(words) = %d, want 1", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew = %q, want %q", words[0].Hebrew, "שלום")
	}
}

// --- ParseWordList: TSV ---

func TestParseWordList_TSVThreeColumns(t *testing.T) {
	input := "שלום\tshalom\tpeace\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("len(words) = %d, want 1", len(words))
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("Transliteration = %q, want %q", words[0].Transliteration, "shalom")
	}
	if words[0].Meaning != "peace" {
		t.Errorf("Meaning = %q, want %q", words[0].Meaning, "peace")
	}
}

func TestParseWordList_TSVTwoColumns(t *testing.T) {
	input := "שלום\tshalom\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("len(words) = %d, want 1", len(words))
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("Transliteration = %q, want %q", words[0].Transliteration, "shalom")
	}
	if words[0].Meaning != "" {
		t.Errorf("Meaning = %q, want empty", words[0].Meaning)
	}
}

// --- ParseWordList: mixed plain text and TSV ---

func TestParseWordList_MixedPlainAndTSV(t *testing.T) {
	// One plain-text line and one TSV line; both should be stored.
	input := "אמת\nשלום\tshalom\tpeace\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList(%q): unexpected error: %v", input, err)
	}
	// אמת = aleph(1)+mem(40)+tav(400) = 441 in Hechrachi
	emet, _, err := src.FindByValue(441, Hechrachi, 10)
	if err != nil {
		t.Fatalf("ParseWordList(%q): FindByValue(441) error: %v", input, err)
	}
	if len(emet) != 1 {
		t.Errorf("ParseWordList(%q): FindByValue(441) got %d words, want 1", input, len(emet))
	}
	// שלום = 376 in Hechrachi
	shalom, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("ParseWordList(%q): FindByValue(376) error: %v", input, err)
	}
	if len(shalom) != 1 {
		t.Errorf("ParseWordList(%q): FindByValue(376) got %d words, want 1", input, len(shalom))
	}
	if shalom[0].Transliteration != "shalom" {
		t.Errorf("ParseWordList(%q): Transliteration = %q, want %q", input, shalom[0].Transliteration, "shalom")
	}
}

// --- ParseWordList: extra tabs beyond three columns ---

func TestParseWordList_ExtraTabsIgnored(t *testing.T) {
	// Four tab-separated fields — only the first three are used.
	input := "שלום\tshalom\tpeace\textra\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList(%q): unexpected error: %v", input, err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("ParseWordList(%q): FindByValue error: %v", input, err)
	}
	if len(words) != 1 {
		t.Fatalf("ParseWordList(%q): got %d words, want 1", input, len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("ParseWordList(%q): Hebrew = %q, want %q", input, words[0].Hebrew, "שלום")
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("ParseWordList(%q): Transliteration = %q, want %q", input, words[0].Transliteration, "shalom")
	}
	if words[0].Meaning != "peace" {
		t.Errorf("ParseWordList(%q): Meaning = %q, want %q", input, words[0].Meaning, "peace")
	}
}

// --- ParseWordList: skip blank lines and comments ---

func TestParseWordList_SkipsBlankLines(t *testing.T) {
	input := "\n\nשלום\n\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 1 {
		t.Errorf("len(words) = %d, want 1", len(words))
	}
}

func TestParseWordList_SkipsCommentLines(t *testing.T) {
	input := "# this is a comment\nשלום\n# another comment\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 1 {
		t.Errorf("len(words) = %d, want 1; comments should be excluded", len(words))
	}
}

// --- FindByValue: different systems produce different results ---

func TestFindByValue_DifferentSystems(t *testing.T) {
	// ך (kaf sofit) = 20 in Hechrachi, 500 in Gadol.
	// A word list with a single ך should be found by value 20 in Hechrachi
	// but not by value 20 in Gadol (it has value 500 in Gadol).
	src, err := ParseWordList(strings.NewReader("ך\n"))
	if err != nil {
		t.Fatalf("ParseWordList: unexpected error: %v", err)
	}

	// Hechrachi: ך = 20 → should match
	hechiWords, _, err := src.FindByValue(20, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue(20, Hechrachi): error: %v", err)
	}
	if len(hechiWords) != 1 {
		t.Errorf("FindByValue(20, Hechrachi): got %d words, want 1", len(hechiWords))
	}

	// Gadol: ך = 500 → value 20 should not match
	gadolMiss, _, err := src.FindByValue(20, Gadol, 10)
	if err != nil {
		t.Fatalf("FindByValue(20, Gadol): error: %v", err)
	}
	if len(gadolMiss) != 0 {
		t.Errorf("FindByValue(20, Gadol): got %d words, want 0 (sofit value is 500 in Gadol)", len(gadolMiss))
	}

	// Gadol: ך = 500 → should match at value 500
	gadolHit, _, err := src.FindByValue(500, Gadol, 10)
	if err != nil {
		t.Fatalf("FindByValue(500, Gadol): error: %v", err)
	}
	if len(gadolHit) != 1 {
		t.Errorf("FindByValue(500, Gadol): got %d words, want 1", len(gadolHit))
	}
}

// --- ParseWordList: no match ---

func TestParseWordList_NoMatchReturnsEmpty(t *testing.T) {
	src, err := ParseWordList(strings.NewReader("שלום\n"))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, hasMore, err := src.FindByValue(999, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if hasMore {
		t.Error("hasMore = true, want false")
	}
	if len(words) != 0 {
		t.Errorf("len(words) = %d, want 0", len(words))
	}
}

// --- FindByValue: invalid Hebrew words are silently skipped ---

func TestFindByValue_SkipsInvalidHebrew(t *testing.T) {
	// "notHebrew" is plain Latin text — Compute will fail for it
	input := "notHebrew\nשלום\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, _, err := src.FindByValue(376, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	// Only שלום (376) should match; notHebrew has no computable value
	if len(words) != 1 {
		t.Errorf("len(words) = %d, want 1 (invalid word silently skipped)", len(words))
	}
}

// --- FindByValue: limit and hasMore ---

func TestFindByValue_LimitEnforced(t *testing.T) {
	// Three words all with value 1 (aleph = 1 in Hechrachi)
	input := "א\nא\nא\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, hasMore, err := src.FindByValue(1, Hechrachi, 2)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 2 {
		t.Errorf("len(words) = %d, want 2 (limited)", len(words))
	}
	if !hasMore {
		t.Error("hasMore = false, want true (third match exists beyond limit)")
	}
}

func TestFindByValue_ExactlyLimitResultsNoHasMore(t *testing.T) {
	// Two words, limit=2 — no overflow
	input := "א\nא\n"
	src, err := ParseWordList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, hasMore, err := src.FindByValue(1, Hechrachi, 2)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 2 {
		t.Errorf("len(words) = %d, want 2", len(words))
	}
	if hasMore {
		t.Error("hasMore = true, want false (exactly limit results, none beyond)")
	}
}

// --- ParseWordList: empty input ---

func TestParseWordList_EmptyReader(t *testing.T) {
	src, err := ParseWordList(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ParseWordList returned error: %v", err)
	}
	words, hasMore, err := src.FindByValue(1, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue returned error: %v", err)
	}
	if len(words) != 0 {
		t.Errorf("len(words) = %d, want 0", len(words))
	}
	if hasMore {
		t.Error("hasMore = true, want false")
	}
}

// --- ParseWordList: reader error propagation ---

func TestParseWordList_ReaderError(t *testing.T) {
	_, err := ParseWordList(&errorReader{})
	if err == nil {
		t.Fatal("ParseWordList returned nil error for failing reader")
	}
}

// --- Package-level FindByValue ---

func TestFindByValue_NilSourceReturnsError(t *testing.T) {
	_, _, err := FindByValue(1, nil, Hechrachi, 10)
	if err == nil {
		t.Fatal("FindByValue(nil source) returned nil error, want non-nil")
	}
}

func TestFindByValue_DefaultLimitConstantIs20(t *testing.T) {
	if DefaultLookupLimit != 20 {
		t.Errorf("DefaultLookupLimit = %d, want 20", DefaultLookupLimit)
	}
}

func TestFindByValue_ZeroLimitUsesDefault(t *testing.T) {
	// Build a word list with DefaultLookupLimit+1 aleph entries (value 1 each).
	// With limit=0 (normalised to DefaultLookupLimit), we expect exactly
	// DefaultLookupLimit results and hasMore=true.
	var sb strings.Builder
	for i := 0; i < DefaultLookupLimit+1; i++ {
		sb.WriteString("א\n")
	}
	src, err := ParseWordList(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("ParseWordList error: %v", err)
	}
	words, hasMore, err := FindByValue(1, src, Hechrachi, 0)
	if err != nil {
		t.Fatalf("FindByValue error: %v", err)
	}
	if len(words) != DefaultLookupLimit {
		t.Errorf("len(words) = %d, want %d", len(words), DefaultLookupLimit)
	}
	if !hasMore {
		t.Error("hasMore = false, want true")
	}
}

func TestFindByValue_NegativeLimitUsesDefault(t *testing.T) {
	src, err := ParseWordList(strings.NewReader("א\n"))
	if err != nil {
		t.Fatalf("ParseWordList error: %v", err)
	}
	// Negative limit should not error and should use DefaultLookupLimit.
	words, _, err := FindByValue(1, src, Hechrachi, -5)
	if err != nil {
		t.Fatalf("FindByValue error: %v", err)
	}
	if len(words) != 1 {
		t.Errorf("len(words) = %d, want 1", len(words))
	}
}

func TestFindByValue_PositiveLimitPassedThrough(t *testing.T) {
	// Three aleph entries; limit=2 should return 2 with hasMore=true.
	src, err := ParseWordList(strings.NewReader("א\nא\nא\n"))
	if err != nil {
		t.Fatalf("ParseWordList error: %v", err)
	}
	words, hasMore, err := FindByValue(1, src, Hechrachi, 2)
	if err != nil {
		t.Fatalf("FindByValue error: %v", err)
	}
	if len(words) != 2 {
		t.Errorf("len(words) = %d, want 2", len(words))
	}
	if !hasMore {
		t.Error("hasMore = false, want true")
	}
}

func TestFindByValue_DelegatesToSource(t *testing.T) {
	// Verify FindByValue delegates correctly: שלום = 376 in Hechrachi.
	src, err := ParseWordList(strings.NewReader("שלום\tshalom\tpeace\n"))
	if err != nil {
		t.Fatalf("ParseWordList error: %v", err)
	}
	words, hasMore, err := FindByValue(376, src, Hechrachi, 10)
	if err != nil {
		t.Fatalf("FindByValue error: %v", err)
	}
	if hasMore {
		t.Error("hasMore = true, want false")
	}
	if len(words) != 1 {
		t.Fatalf("len(words) = %d, want 1", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew = %q, want %q", words[0].Hebrew, "שלום")
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("Transliteration = %q, want %q", words[0].Transliteration, "shalom")
	}
}

// errorReader always returns an error on Read.
type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}
