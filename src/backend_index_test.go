package gematria_test

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

// indexContent builds a minimal valid index file as a string.
// Lines must be pre-sorted (system lex, value numeric).
func indexContent(lines ...string) string {
	const header = "# gematria-index v1\n"
	if len(lines) == 0 {
		return header
	}
	return header + strings.Join(lines, "\n") + "\n"
}

// TestNewIndexWordSource_FindByValue_match is the tracer bullet:
// an index with one matching word returns that word via FindByValue.
func TestNewIndexWordSource_FindByValue_match(t *testing.T) {
	content := indexContent("hechrachi\t376\tשלום\tshalom\tpeace")
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew: got %q, want %q", words[0].Hebrew, "שלום")
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("Transliteration: got %q, want %q", words[0].Transliteration, "shalom")
	}
	if words[0].Meaning != "peace" {
		t.Errorf("Meaning: got %q, want %q", words[0].Meaning, "peace")
	}
}

// TestNewIndexWordSource_FindByValue_noMatch verifies no results when nothing matches.
func TestNewIndexWordSource_FindByValue_noMatch(t *testing.T) {
	content := indexContent("hechrachi\t376\tשלום\tshalom\tpeace")
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(999, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

// TestNewIndexWordSource_FindByValue_hasMore verifies hasMore=true when
// results exceed the limit.
func TestNewIndexWordSource_FindByValue_hasMore(t *testing.T) {
	content := indexContent(
		"hechrachi\t441\tאמת\t\t",
		"hechrachi\t441\tתמא\t\t",
		"hechrachi\t441\tמתא\t\t",
	)
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(441, gematria.Hechrachi, 2)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if !hasMore {
		t.Error("hasMore should be true when results exceed limit")
	}
	if len(words) != 2 {
		t.Errorf("expected 2 words (limit), got %d", len(words))
	}
}

// TestNewIndexWordSource_FindByValue_systemFilter verifies that only words
// for the requested system are returned.
func TestNewIndexWordSource_FindByValue_systemFilter(t *testing.T) {
	content := indexContent(
		"gadol\t999\ttest\t\t",
		"hechrachi\t376\tשלום\tshalom\tpeace",
	)
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, _, err := src.FindByValue(999, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words for wrong system, got %d", len(words))
	}
}

// TestNewIndexWordSource_FindByValue_missingOptionalFields verifies that
// a word line with only hebrew (no transliteration/meaning) parses correctly.
func TestNewIndexWordSource_FindByValue_missingOptionalFields(t *testing.T) {
	content := indexContent("hechrachi\t376\tשלום")
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, _, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew: got %q, want %q", words[0].Hebrew, "שלום")
	}
	if words[0].Transliteration != "" {
		t.Errorf("Transliteration: got %q, want empty", words[0].Transliteration)
	}
	if words[0].Meaning != "" {
		t.Errorf("Meaning: got %q, want empty", words[0].Meaning)
	}
}

// TestNewIndexWordSource_missingMagicComment verifies that a file without
// the required magic comment header returns an error.
func TestNewIndexWordSource_missingMagicComment(t *testing.T) {
	content := "hechrachi\t376\tשלום\tshalom\tpeace\n"
	r := strings.NewReader(content)

	_, err := gematria.NewIndexWordSource(r)
	if err == nil {
		t.Fatal("expected error for file without magic comment, got nil")
	}
}

// TestNewIndexWordSource_commentsAndBlanksSkipped verifies that blank lines
// and '#' comment lines in the data section do not interfere with indexing.
func TestNewIndexWordSource_commentsAndBlanksSkipped(t *testing.T) {
	content := "# gematria-index v1\n" +
		"\n" +
		"# this is a comment\n" +
		"hechrachi\t376\tשלום\tshalom\tpeace\n" +
		"\n"
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, _, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if len(words) != 1 {
		t.Errorf("expected 1 word, got %d", len(words))
	}
}

// TestNewIndexWordSource_emptyFile verifies that an empty index (header only)
// returns no results without error.
func TestNewIndexWordSource_emptyFile(t *testing.T) {
	content := "# gematria-index v1\n"
	r := strings.NewReader(content)

	src, err := gematria.NewIndexWordSource(r)
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false for empty index")
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

// TestWriteIndexFile_idempotent verifies that writing the same words twice
// (simulating an additive merge) produces no duplicates. WriteIndexFile deduplicates
// by Hebrew, so appending old+new with overlaps yields the same index.
func TestWriteIndexFile_idempotent(t *testing.T) {
	words := []gematria.Word{
		{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"},
		{Hebrew: "אמת", Transliteration: "emet", Meaning: "truth"},
	}

	// First write.
	var buf1 bytes.Buffer
	count1, err := gematria.WriteIndexFile(&buf1, words)
	if err != nil {
		t.Fatalf("first WriteIndexFile: %v", err)
	}
	if count1 != 2 {
		t.Fatalf("first call: expected count=2, got %d", count1)
	}

	// Read back existing words and merge with the same input (simulates additive).
	existing, err := gematria.ReadIndexWords(bytes.NewReader(buf1.Bytes()))
	if err != nil {
		t.Fatalf("ReadIndexWords: %v", err)
	}
	merged := append(existing, words...)

	var buf2 bytes.Buffer
	count2, err := gematria.WriteIndexFile(&buf2, merged)
	if err != nil {
		t.Fatalf("second WriteIndexFile: %v", err)
	}
	if count2 != 2 {
		t.Fatalf("second call: expected count=2 (same words), got %d", count2)
	}

	// Both outputs should be identical.
	if buf1.String() != buf2.String() {
		t.Error("index file content differs after idempotent write")
	}

	// Verify lookup returns exactly 1 result for שלום.
	src, err := gematria.NewIndexWordSource(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}
	results, hasMore, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for שלום, got %d", len(results))
	}
}

// TestNewIndexWordSource_FindByValue_MatchesInMemory is the oracle test:
// for the same input data, IndexWordSource must return identical results
// to the in-memory ParseWordList for every (value, system) query.
func TestNewIndexWordSource_FindByValue_MatchesInMemory(t *testing.T) {
	const wordlistTSV = "שלום\tshalom\tpeace\n" +
		"אמת\temet\ttruth\n" +
		"אור\tor\tlight\n" +
		"דרך\tderekh\tway\n"

	// Build in-memory source.
	memSrc, err := gematria.ParseWordList(strings.NewReader(wordlistTSV))
	if err != nil {
		t.Fatalf("ParseWordList: %v", err)
	}

	// Build index source from the same words.
	words, err := gematria.ParseWordListSlice(strings.NewReader(wordlistTSV))
	if err != nil {
		t.Fatalf("ParseWordListSlice: %v", err)
	}
	var buf bytes.Buffer
	if _, err := gematria.WriteIndexFile(&buf, words); err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}
	idxSrc, err := gematria.NewIndexWordSource(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	// Test several (value, system) pairs covering matches, no-match, and multiple systems.
	cases := []struct {
		value  int
		system gematria.System
	}{
		{376, gematria.Hechrachi}, // שלום
		{441, gematria.Hechrachi}, // אמת
		{207, gematria.Hechrachi}, // אור
		{224, gematria.Hechrachi}, // דרך
		{999, gematria.Hechrachi}, // no match
		{376, gematria.Gadol},     // שלום in gadol
		{376, gematria.Siduri},    // שלום in siduri
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("value=%d_system=%s", tc.value, tc.system), func(t *testing.T) {
			memWords, memHasMore, err := memSrc.FindByValue(tc.value, tc.system, 20)
			if err != nil {
				t.Fatalf("memory FindByValue: %v", err)
			}
			idxWords, idxHasMore, err := idxSrc.FindByValue(tc.value, tc.system, 20)
			if err != nil {
				t.Fatalf("index FindByValue: %v", err)
			}

			if memHasMore != idxHasMore {
				t.Errorf("hasMore: memory=%v index=%v", memHasMore, idxHasMore)
			}
			if len(memWords) != len(idxWords) {
				t.Errorf("count: memory=%d index=%d", len(memWords), len(idxWords))
				return
			}
			// Sort by Hebrew before comparing to tolerate any ordering differences.
			sort.Slice(memWords, func(i, j int) bool { return memWords[i].Hebrew < memWords[j].Hebrew })
			sort.Slice(idxWords, func(i, j int) bool { return idxWords[i].Hebrew < idxWords[j].Hebrew })
			for i := range memWords {
				if memWords[i].Hebrew != idxWords[i].Hebrew {
					t.Errorf("[%d] Hebrew: memory=%q index=%q", i, memWords[i].Hebrew, idxWords[i].Hebrew)
				}
				if memWords[i].Transliteration != idxWords[i].Transliteration {
					t.Errorf("[%d] Transliteration: memory=%q index=%q", i, memWords[i].Transliteration, idxWords[i].Transliteration)
				}
				if memWords[i].Meaning != idxWords[i].Meaning {
					t.Errorf("[%d] Meaning: memory=%q index=%q", i, memWords[i].Meaning, idxWords[i].Meaning)
				}
			}
		})
	}
}

// TestNewIndexWordSource_FindByValue_BoundaryValues verifies that words with
// the minimum and maximum single-letter hechrachi values are found correctly.
func TestNewIndexWordSource_FindByValue_BoundaryValues(t *testing.T) {
	// Aleph (א) = 1 in hechrachi — minimum single-letter value.
	// Tav (ת) = 400 in hechrachi — maximum single-letter value.
	const wordlistTSV = "א\taleph\toxen\nת\ttav\tcross\n"

	words, err := gematria.ParseWordListSlice(strings.NewReader(wordlistTSV))
	if err != nil {
		t.Fatalf("ParseWordListSlice: %v", err)
	}
	var buf bytes.Buffer
	if _, err := gematria.WriteIndexFile(&buf, words); err != nil {
		t.Fatalf("WriteIndexFile: %v", err)
	}
	src, err := gematria.NewIndexWordSource(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("NewIndexWordSource: %v", err)
	}

	// Minimum: aleph = 1
	got, hasMore, err := src.FindByValue(1, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue(1): %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false for minimum-value word")
	}
	if len(got) != 1 {
		t.Fatalf("FindByValue(1): expected 1 word, got %d", len(got))
	}
	if got[0].Hebrew != "א" {
		t.Errorf("FindByValue(1): Hebrew = %q, want %q", got[0].Hebrew, "א")
	}

	// Maximum single-letter: tav = 400
	got, hasMore, err = src.FindByValue(400, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue(400): %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false for maximum-value word")
	}
	if len(got) != 1 {
		t.Fatalf("FindByValue(400): expected 1 word, got %d", len(got))
	}
	if got[0].Hebrew != "ת" {
		t.Errorf("FindByValue(400): Hebrew = %q, want %q", got[0].Hebrew, "ת")
	}
}
