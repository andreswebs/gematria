---
id: wor-22g6
status: closed
deps: [wor-alrj, wor-9xcu, wor-0153, wor-92gu]
links: []
created: 2026-04-15T04:18:05Z
type: task
priority: 2
parent: wor-6uug
tags: [wordlist-backends, backend-selection, task]
---
# Backend auto-selection and io.Closer integration in CLI

Modify src/internal/cli/run.go to detect which WordSource backend to use based on the --wordlist path or URL, construct the appropriate backend, and safely clean up backends that implement io.Closer. Also adds an optional --wordlist-format flag for explicit backend override and the GEMATRIA_WORDLIST_TOKEN env var for remote authentication.

## Spec References
- docs/specs/wordlist-backends.md §"Suggestions" (backend selection belongs in internal/cli)
- docs/specs/cli-design.md §"Environment Variables" (GEMATRIA_WORDLIST pattern; lazy validation)
- docs/specs/code-architecture.md §"CLI layer" (Run signature, Config struct, lazy env var validation)

## Implementation Details

### openWordSource Helper

Add a private function in src/internal/cli/run.go (or a new file src/internal/cli/backends.go):

  func openWordSource(
    path string,
    getenv func(string) string,
  ) (gematria.WordSource, io.Closer, error)

Returns the WordSource and optionally an io.Closer (nil when the backend needs no cleanup). Caller always calls:
  if closer != nil { defer closer.Close() }

Detection order (evaluated top to bottom, first match wins):
  1. Explicit --wordlist-format flag present → use the specified backend regardless of path shape
  2. path starts with "http://" or "https://" → remote backend
       token := getenv("GEMATRIA_WORDLIST_TOKEN")
       opts := []gematria.RemoteOption{}
       if token != "" { opts = append(opts, gematria.WithAuthToken(token)) }
       src, err := gematria.NewRemoteWordSource(path, opts...)
       return src, nil, err   // no Closer needed
  3. path ends with ".db" → SQLite backend
       src, err := gematria.OpenSQLiteWordSource(path)
       return src, src.(io.Closer), err   // sqliteWordSource implements io.Closer
  4. path ends with ".idx" → index file backend
       f, err := os.Open(path)
       src, err := gematria.NewIndexWordSource(f)
       return src, f, err   // caller closes the *os.File
  5. companion file path+".idx" exists (os.Stat) → index file backend with companion
       same as case 4 but with path = path+".idx"
  6. default → in-memory (ParseWordList)
       f, err := os.Open(path)
       src, err := gematria.ParseWordList(f)
       f.Close()             // safe to close immediately; data already in memory
       return src, nil, err

### Config Additions (src/internal/cli/config.go)

New fields:
  WordlistFormat string // --wordlist-format flag; empty means auto-detect

New flag registered in the pflag set:
  --wordlist-format string  Override backend auto-detection (sqlite|index|remote|memory)

New env var (lazy, only accessed when remote backend is selected):
  GEMATRIA_WORDLIST_TOKEN   Bearer token for authenticated remote word sources

Validation: if --wordlist-format is non-empty and not one of (sqlite, index, remote, memory), print error listing valid values and return exit code 2.

### Integration into Run()

Replace the current inline os.Open + ParseWordList call in the --find branch with:
  src, closer, err := openWordSource(cfg.Wordlist, getenv)
  if err != nil {
    // format and print error to stderr; return 3 for file errors, 1 for domain errors
  }
  if closer != nil { defer closer.Close() }
  // proceed with gematria.FindByValue(cfg.FindValue, src, cfg.System, cfg.Limit)


## Notes

**2026-04-15T13:16:26Z**

Implemented openWordSource helper in run.go with auto-detection by path shape (http:// → remote, .db → SQLite, .idx → index, companion .idx → index, default → memory). Added WordlistFormat field to Config and --wordlist-format flag (sqlite|index|remote|memory) for explicit override. GEMATRIA_WORDLIST_TOKEN env var wires into remote backend's WithAuthToken option. runFind now calls openWordSource and defers closer.Close() when the backend returns a non-nil io.Closer. Covered by 11 new integration tests in run_backends_test.go using real temp files and httptest.NewServer for remote.
