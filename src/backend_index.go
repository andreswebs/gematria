package gematria

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// indexMagic is the required first non-blank line of a gematria index file.
const indexMagic = "# gematria-index v1"

// indexRange records the byte offset of the first line and the total count of
// words for a single (system, value) group in an index file. Storing only
// offsets and counts keeps the jump table memory cost proportional to the
// number of distinct (system, value) pairs, not the number of words.
type indexRange struct {
	offset int64
	count  int
}

// indexWordSource is an unexported WordSource backed by a pre-computed sorted
// TSV index file produced by "gematria index --format index".
type indexWordSource struct {
	rs    io.ReadSeeker
	index map[System]map[int]indexRange
}

// Compile-time assertion that *indexWordSource satisfies WordSource.
var _ WordSource = (*indexWordSource)(nil)

// NewIndexWordSource builds an in-memory jump table from r by scanning the
// index file once, then returns a WordSource whose FindByValue performs an
// O(1) jump-table lookup followed by a sequential read.
//
// r must be positioned at the beginning of a file produced by
// "gematria index --format index". The caller is responsible for closing r
// if it implements io.Closer.
func NewIndexWordSource(r io.ReadSeeker) (WordSource, error) {
	src := &indexWordSource{
		rs:    r,
		index: make(map[System]map[int]indexRange),
	}
	if err := src.buildIndex(); err != nil {
		return nil, err
	}
	return src, nil
}

// buildIndex performs a single sequential scan of the index file, validates
// the magic comment, and populates the in-memory jump table.
func (s *indexWordSource) buildIndex() error {
	if _, err := s.rs.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("gematria: index seek: %w", err)
	}

	br := bufio.NewReader(s.rs)
	var pos int64

	// Validate magic comment — must appear as the first non-blank line.
	magicFound := false
	for {
		line, err := br.ReadString('\n')
		pos += int64(len(line))
		trimmed := strings.TrimRight(line, "\r\n")

		if err != nil && err != io.EOF {
			return fmt.Errorf("gematria: index read: %w", err)
		}
		if trimmed != "" {
			if trimmed != indexMagic {
				return fmt.Errorf("gematria: index: missing or invalid magic comment (want %q)", indexMagic)
			}
			magicFound = true
			break
		}
		if err == io.EOF {
			break
		}
	}
	if !magicFound {
		return fmt.Errorf("gematria: index: missing or invalid magic comment (want %q)", indexMagic)
	}

	// Parse data lines and build the jump table. Lines are expected to be
	// sorted by (system, value) so we accumulate consecutive equal pairs.
	var (
		curSystem System
		curValue  int
		curOffset int64
		curCount  int
		started   bool
	)

	flush := func() {
		if !started || curCount == 0 {
			return
		}
		if s.index[curSystem] == nil {
			s.index[curSystem] = make(map[int]indexRange)
		}
		s.index[curSystem][curValue] = indexRange{offset: curOffset, count: curCount}
	}

	for {
		startPos := pos
		line, err := br.ReadString('\n')
		pos += int64(len(line))
		trimmed := strings.TrimRight(line, "\r\n")

		if err != nil && err != io.EOF {
			flush()
			return fmt.Errorf("gematria: index read: %w", err)
		}

		// Skip blank lines and comment lines.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if err == io.EOF {
				break
			}
			continue
		}

		// Expect at least system, value, hebrew (3 tab-separated fields).
		parts := strings.SplitN(trimmed, "\t", 4)
		if len(parts) < 3 {
			if err == io.EOF {
				break
			}
			continue
		}

		system := System(parts[0])
		value, parseErr := strconv.Atoi(parts[1])
		if parseErr != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if !started || system != curSystem || value != curValue {
			flush()
			curSystem = system
			curValue = value
			curOffset = startPos
			curCount = 1
			started = true
		} else {
			curCount++
		}

		if err == io.EOF {
			break
		}
	}
	flush()

	return nil
}

// FindByValue seeks to the pre-computed offset for (system, value) and reads
// up to limit word records. Returns (nil, false, nil) when no words match.
func (s *indexWordSource) FindByValue(value int, system System, limit int) ([]Word, bool, error) {
	sysIndex, ok := s.index[system]
	if !ok {
		return nil, false, nil
	}
	r, ok := sysIndex[value]
	if !ok {
		return nil, false, nil
	}

	hasMore := r.count > limit
	toRead := r.count
	if hasMore {
		toRead = limit
	}

	if _, err := s.rs.Seek(r.offset, io.SeekStart); err != nil {
		return nil, false, fmt.Errorf("gematria: index seek: %w", err)
	}

	br := bufio.NewReader(s.rs)
	results := make([]Word, 0, toRead)

	for len(results) < toRead {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, false, fmt.Errorf("gematria: index read: %w", err)
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			results = append(results, parseIndexLine(trimmed))
		}
		if err == io.EOF {
			break
		}
	}

	return results, hasMore, nil
}

// parseIndexLine extracts a Word from a data line in the index format:
// <system>\t<value>\t<hebrew>[\t<transliteration>[\t<meaning>]]
func parseIndexLine(line string) Word {
	parts := strings.Split(line, "\t")
	var w Word
	if len(parts) >= 3 {
		w.Hebrew = parts[2]
	}
	if len(parts) > 3 {
		w.Transliteration = parts[3]
	}
	if len(parts) > 4 {
		w.Meaning = parts[4]
	}
	return w
}

// ReadIndexWords reads an index file and returns the unique words it contains.
// This is useful for merging an existing index with new words.
func ReadIndexWords(r io.Reader) ([]Word, error) {
	br := bufio.NewReader(r)
	seen := make(map[string]bool)
	var words []Word

	for {
		line, err := br.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			w := parseIndexLine(trimmed)
			if w.Hebrew != "" && !seen[w.Hebrew] {
				seen[w.Hebrew] = true
				words = append(words, w)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gematria: read index: %w", err)
		}
	}
	return words, nil
}

// indexEntry is an intermediate record used when building an index file in memory.
type indexEntry struct {
	system System
	value  int
	word   Word
}

// WriteIndexFile writes a pre-computed index to w in the format expected by
// NewIndexWordSource. It computes gematria values for all four systems for each
// word, sorts entries by (system, value), and writes them with the magic header.
//
// Words whose Hebrew text fails Compute for a given system are silently skipped
// for that system. Duplicate Hebrew words are silently skipped (idempotent).
//
// Returns (count, error) where count is the number of new distinct words indexed
// (words that produced at least one valid value entry).
func WriteIndexFile(w io.Writer, words []Word) (int, error) {
	systems := ValidSystems()

	var entries []indexEntry
	seen := map[string]bool{} // deduplicate by Hebrew

	for _, word := range words {
		if seen[word.Hebrew] {
			continue
		}
		any := false
		for _, sys := range systems {
			result, err := Compute(word.Hebrew, sys)
			if err != nil {
				continue
			}
			entries = append(entries, indexEntry{
				system: sys,
				value:  result.Total,
				word:   word,
			})
			any = true
		}
		if any {
			seen[word.Hebrew] = true
		}
	}

	// Sort by system name then value (numeric).
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].system != entries[j].system {
			return string(entries[i].system) < string(entries[j].system)
		}
		return entries[i].value < entries[j].value
	})

	bw := bufio.NewWriter(w)
	if _, err := fmt.Fprintln(bw, indexMagic); err != nil {
		return 0, err
	}
	for _, e := range entries {
		line := fmt.Sprintf("%s\t%d\t%s\t%s\t%s",
			string(e.system),
			e.value,
			e.word.Hebrew,
			e.word.Transliteration,
			e.word.Meaning,
		)
		if _, err := fmt.Fprintln(bw, line); err != nil {
			return 0, err
		}
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	return len(seen), nil
}
