package gematria

import (
	"bufio"
	"io"
	"strings"
)

// Word is a single entry from a word list used for reverse lookups.
// Hebrew is required. Transliteration and Meaning are optional and come
// from the second and third tab-separated columns of a TSV word list.
type Word struct {
	Hebrew          string
	Transliteration string
	Meaning         string
}

// WordSource is the interface for reverse-lookup word-list backends.
// Three implementations exist: in-memory (ParseWordList), SQLite
// (OpenSQLiteWordSource), and pre-computed index (NewIndexWordSource).
// Backends that need cleanup implement io.Closer separately.
//
// FindByValue returns at most limit Words whose gematria value in system
// equals value. hasMore is true when additional matching words exist
// beyond the returned slice. Returns (nil, false, nil) when no words match.
type WordSource interface {
	FindByValue(value int, system System, limit int) ([]Word, bool, error)
}

// wordList is the in-memory WordSource implementation returned by ParseWordList.
type wordList struct {
	words []Word
}

// Compile-time assertion that *wordList satisfies WordSource.
var _ WordSource = (*wordList)(nil)

// ParseWordList parses a word list from r and returns an in-memory WordSource.
// It accepts two formats:
//   - Plain text: one Hebrew word per line.
//   - TSV: word[TAB]transliteration[TAB]meaning (transliteration and meaning optional).
//
// Blank lines and lines beginning with '#' are silently ignored.
// Returns an error only if reading from r fails.
func ParseWordList(r io.Reader) (WordSource, error) {
	words, err := ParseWordListSlice(r)
	if err != nil {
		return nil, err
	}
	return &wordList{words: words}, nil
}

// ParseWordListSlice parses a word list from r and returns the words as a slice.
// It uses the same parsing rules as ParseWordList.
// This is provided for callers (such as the index subcommand) that need to
// iterate all words directly rather than through the WordSource interface.
func ParseWordListSlice(r io.Reader) ([]Word, error) {
	scanner := bufio.NewScanner(r)
	var words []Word
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var w Word
		if strings.ContainsRune(line, '\t') {
			parts := strings.Split(line, "\t")
			w.Hebrew = parts[0]
			if len(parts) > 1 {
				w.Transliteration = parts[1]
			}
			if len(parts) > 2 {
				w.Meaning = parts[2]
			}
		} else {
			w.Hebrew = line
		}
		words = append(words, w)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

// FindByValue returns at most limit Words from the list whose gematria value
// under system equals value. hasMore is true when additional matches exist
// beyond the returned slice. Words where Compute fails are silently skipped.
func (wl *wordList) FindByValue(value int, system System, limit int) ([]Word, bool, error) {
	var matches []Word
	for _, w := range wl.words {
		result, err := Compute(w.Hebrew, system)
		if err != nil {
			continue
		}
		if result.Total == value {
			matches = append(matches, w)
			if len(matches) == limit+1 {
				return matches[:limit], true, nil
			}
		}
	}
	return matches, false, nil
}
