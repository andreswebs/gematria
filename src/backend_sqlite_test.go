package gematria_test

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
	_ "modernc.org/sqlite"
)

// createTestDB creates a temporary SQLite database with the gematria schema
// and returns its path. The caller is responsible for removing the file.
func createTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("createTestDB: sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`
		CREATE TABLE words (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hebrew TEXT NOT NULL,
			transliteration TEXT NOT NULL DEFAULT '',
			meaning TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE word_values (
			word_id INTEGER NOT NULL REFERENCES words(id),
			system TEXT NOT NULL,
			value INTEGER NOT NULL
		);
		CREATE INDEX idx_word_values ON word_values(system, value);
	`)
	if err != nil {
		t.Fatalf("createTestDB: schema: %v", err)
	}
	return path
}

// insertWord inserts a word and its value into the test database.
func insertWord(t *testing.T, db *sql.DB, hebrew, transliteration, meaning, system string, value int) {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO words (hebrew, transliteration, meaning) VALUES (?, ?, ?)`,
		hebrew, transliteration, meaning,
	)
	if err != nil {
		t.Fatalf("insertWord: %v", err)
	}
	id, _ := res.LastInsertId()
	_, err = db.Exec(
		`INSERT INTO word_values (word_id, system, value) VALUES (?, ?, ?)`,
		id, system, value,
	)
	if err != nil {
		t.Fatalf("insertWord word_values: %v", err)
	}
}

// TestOpenSQLiteWordSource_FindByValue_match is the tracer bullet:
// a DB with one matching word returns that word via FindByValue.
func TestOpenSQLiteWordSource_FindByValue_match(t *testing.T) {
	path := createTestDB(t)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	insertWord(t, db, "שלום", "shalom", "peace", "hechrachi", 376)
	_ = db.Close()

	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	if c, ok := src.(io.Closer); ok {
		defer func() { _ = c.Close() }()
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

// TestOpenSQLiteWordSource_FindByValue_noMatch verifies that no results and
// hasMore=false are returned when no words match the query.
func TestOpenSQLiteWordSource_FindByValue_noMatch(t *testing.T) {
	path := createTestDB(t)

	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	if c, ok := src.(io.Closer); ok {
		defer func() { _ = c.Close() }()
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

// TestOpenSQLiteWordSource_notFound verifies that opening a non-existent file
// returns an error.
func TestOpenSQLiteWordSource_notFound(t *testing.T) {
	_, err := gematria.OpenSQLiteWordSource("/nonexistent/path/words.db")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

// TestOpenSQLiteWordSource_wrongSchema verifies that opening a file without
// the expected schema returns a clear error.
func TestOpenSQLiteWordSource_wrongSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE unrelated (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_ = db.Close()

	_, err = gematria.OpenSQLiteWordSource(path)
	if err == nil {
		t.Fatal("expected error for wrong schema, got nil")
	}
}

// TestOpenSQLiteWordSource_FindByValue_hasMore verifies that hasMore is true
// when matching words exceed the limit.
func TestOpenSQLiteWordSource_FindByValue_hasMore(t *testing.T) {
	path := createTestDB(t)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Insert 3 words with the same value.
	for i, word := range []string{"אמת", "אמת", "אמת"} {
		_ = i
		insertWord(t, db, word, "", "", "hechrachi", 441)
	}
	_ = db.Close()

	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	if c, ok := src.(io.Closer); ok {
		defer func() { _ = c.Close() }()
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

// TestOpenSQLiteWordSource_FindByValue_systemFilter verifies that FindByValue
// filters by system correctly — a word with a value in "gadol" is not returned
// when querying "hechrachi".
func TestOpenSQLiteWordSource_FindByValue_systemFilter(t *testing.T) {
	path := createTestDB(t)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	insertWord(t, db, "testword", "", "", "gadol", 999)
	_ = db.Close()

	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	if c, ok := src.(io.Closer); ok {
		defer func() { _ = c.Close() }()
	}

	words, _, err := src.FindByValue(999, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words for wrong system, got %d", len(words))
	}
}

// TestOpenSQLiteWordSource_Close verifies that Close can be called on the
// returned WordSource via the io.Closer interface.
func TestOpenSQLiteWordSource_Close(t *testing.T) {
	path := createTestDB(t)

	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	c, ok := src.(io.Closer)
	if !ok {
		t.Fatal("OpenSQLiteWordSource result does not implement io.Closer")
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// TestOpenSQLiteWordSource_plainTextFile verifies that opening a plain text
// file (not a SQLite DB) returns a clear error.
func TestOpenSQLiteWordSource_plainTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	if err := os.WriteFile(path, []byte("שלום\nאמת\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := gematria.OpenSQLiteWordSource(path)
	if err == nil {
		t.Fatal("expected error for plain text file, got nil")
	}
}

// TestOpenSQLiteWordSource_FindByValue_MatchesInMemory is the oracle test:
// for the same input data, SQLiteWordSource must return identical results
// to the in-memory ParseWordList for every (value, system) query.
func TestOpenSQLiteWordSource_FindByValue_MatchesInMemory(t *testing.T) {
	const wordlistTSV = "שלום\tshalom\tpeace\n" +
		"אמת\temet\ttruth\n" +
		"אור\tor\tlight\n" +
		"דרך\tderekh\tway\n"

	// Build in-memory source.
	memSrc, err := gematria.ParseWordList(strings.NewReader(wordlistTSV))
	if err != nil {
		t.Fatalf("ParseWordList: %v", err)
	}

	// Build SQLite source from the same words.
	words, err := gematria.ParseWordListSlice(strings.NewReader(wordlistTSV))
	if err != nil {
		t.Fatalf("ParseWordListSlice: %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "words.db")
	if _, err := gematria.WriteIndexSQLite(dbPath, words); err != nil {
		t.Fatalf("WriteIndexSQLite: %v", err)
	}
	sqlSrc, err := gematria.OpenSQLiteWordSource(dbPath)
	if err != nil {
		t.Fatalf("OpenSQLiteWordSource: %v", err)
	}
	if c, ok := sqlSrc.(io.Closer); ok {
		defer func() { _ = c.Close() }()
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
			sqlWords, sqlHasMore, err := sqlSrc.FindByValue(tc.value, tc.system, 20)
			if err != nil {
				t.Fatalf("sqlite FindByValue: %v", err)
			}

			if memHasMore != sqlHasMore {
				t.Errorf("hasMore: memory=%v sqlite=%v", memHasMore, sqlHasMore)
			}
			if len(memWords) != len(sqlWords) {
				t.Errorf("count: memory=%d sqlite=%d", len(memWords), len(sqlWords))
				return
			}
			// Sort by Hebrew before comparing to tolerate any ordering differences.
			sort.Slice(memWords, func(i, j int) bool { return memWords[i].Hebrew < memWords[j].Hebrew })
			sort.Slice(sqlWords, func(i, j int) bool { return sqlWords[i].Hebrew < sqlWords[j].Hebrew })
			for i := range memWords {
				if memWords[i].Hebrew != sqlWords[i].Hebrew {
					t.Errorf("[%d] Hebrew: memory=%q sqlite=%q", i, memWords[i].Hebrew, sqlWords[i].Hebrew)
				}
				if memWords[i].Transliteration != sqlWords[i].Transliteration {
					t.Errorf("[%d] Transliteration: memory=%q sqlite=%q", i, memWords[i].Transliteration, sqlWords[i].Transliteration)
				}
				if memWords[i].Meaning != sqlWords[i].Meaning {
					t.Errorf("[%d] Meaning: memory=%q sqlite=%q", i, memWords[i].Meaning, sqlWords[i].Meaning)
				}
			}
		})
	}
}
