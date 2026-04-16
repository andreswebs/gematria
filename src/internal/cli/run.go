// Package cli implements the command-line interface for gematria.
package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	gematria "github.com/andreswebs/gematria"
	"github.com/spf13/pflag"
)

// cliVersion is overridden at build time via:
//
//	-ldflags "-X github.com/andreswebs/gematria/internal/cli.cliVersion=<tag>"
//
// The default uses a v-prefixed pseudo-version so dev and released builds
// share one canonical format. See .local/docs/go-versioning-v-prefix.md.
var cliVersion = "v0.0.0-dev"

const helpText = `Usage: gematria [OPTIONS] [INPUT...]

Compute the gematria (Hebrew numerology) value of Hebrew words or letters.

Arguments:
  INPUT    Hebrew Unicode text, Latin transliteration, or mixed input.
           When omitted and stdin is not a terminal, reads lines from stdin.

Options:
  -m, --mispar string           Gematria system: hechrachi|gadol|siduri|atbash
                                (default: hechrachi)
  -o, --output string           Output format: line|value|card|json (default: line)
  -t, --transliterate           Interpret Latin input as Hebrew words (per --scheme)
      --scheme string           Transliteration scheme: academic|israeli (default: academic)
      --atbash                  Show Atbash letter substitutions in output
      --no-color                Disable ANSI color output
      --fail-early              Stop on first error when reading from stdin
      --version                 Print version and exit
      --wordlist-format string  Backend override: sqlite|index|remote|memory
  -h, --help                    Show this help message

Environment Variables:
  GEMATRIA_MISPAR          Default gematria system (same values as --mispar)
  GEMATRIA_OUTPUT          Default output format (same values as --output)
  GEMATRIA_SCHEME          Default transliteration scheme (academic|israeli);
                           validated lazily, only when -t is active
  GEMATRIA_WORDLIST        Path to word list file for reverse lookup (--find)
  GEMATRIA_LIMIT           Maximum number of reverse lookup results (default: 20)
  GEMATRIA_WORDLIST_TOKEN  Bearer token for authenticated remote word sources
  NO_COLOR                 Set to any value to disable ANSI color output

Examples:
  gematria א                              Compute aleph (single letter)
  gematria שלום                           Compute shalom (word)
  gematria --mispar gadol שרה             Compute with Gadol system
  gematria --output json שלום             JSON output for scripting
  echo "שלום" | gematria                  Read Hebrew from stdin
  gematria -t shalom                      Transliterate "shalom" (academic scheme)
  gematria -t --scheme israeli gadol      Transliterate "gadol" (israeli scheme)
  gematria --find 376 --wordlist w.txt    Reverse lookup: find words = 376`

const indexHelpText = `Usage: gematria index [OPTIONS]

Generate a pre-computed index from a word list for fast reverse lookups.

Options:
      --wordlist string   Path to input word list (required)
      --output string     Output file path (default: <wordlist>.db or <wordlist>.idx)
      --format string     Output format: sqlite|index (default: sqlite)
  -h, --help              Show this help message

Examples:
  gematria index --wordlist words.txt
  gematria index --wordlist words.txt --format index
  gematria index --wordlist words.txt --output custom.db --format sqlite`

// runIndex implements the "gematria index" subcommand.
// It reads a word list, computes gematria values for all four systems,
// and writes a pre-computed index in the requested format.
func runIndex(args []string, stdout *os.File, stderr *os.File) int {
	fs := pflag.NewFlagSet("gematria index", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var wordlistPath string
	var outputPath string
	var format string
	var help bool

	fs.StringVar(&wordlistPath, "wordlist", "", "path to input word list")
	fs.StringVar(&outputPath, "output", "", "output file path")
	fs.StringVar(&format, "format", "sqlite", "output format: sqlite|index")
	fs.BoolVarP(&help, "help", "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err.Error())
		return 2
	}

	if help {
		_, _ = fmt.Fprintln(stdout, indexHelpText)
		return 0
	}

	if format != "sqlite" && format != "index" {
		_, _ = fmt.Fprintf(stderr, "Error: invalid value %q for --format\nvalid values: sqlite, index\n", format)
		return 2
	}

	if wordlistPath == "" {
		_, _ = fmt.Fprintln(stderr, "Error: --wordlist is required for the index subcommand")
		return 2
	}

	// Determine default output path.
	if outputPath == "" {
		if format == "sqlite" {
			outputPath = wordlistPath + ".db"
		} else {
			outputPath = wordlistPath + ".idx"
		}
	}

	// Open and parse the word list.
	f, err := os.Open(wordlistPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: cannot open word list %q: %v\n", wordlistPath, err)
		return 3
	}
	words, err := gematria.ParseWordListSlice(f)
	_ = f.Close()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: cannot read word list %q: %v\n", wordlistPath, err)
		return 3
	}

	// Write output in the requested format.
	var count int
	switch format {
	case "sqlite":
		count, err = gematria.WriteIndexSQLite(outputPath, words)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: cannot write index %q: %v\n", outputPath, err)
			return 3
		}
	case "index":
		out, ferr := os.Create(outputPath)
		if ferr != nil {
			_, _ = fmt.Fprintf(stderr, "Error: cannot create output file %q: %v\n", outputPath, ferr)
			return 3
		}
		count, err = gematria.WriteIndexFile(out, words)
		_ = out.Close()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: cannot write index %q: %v\n", outputPath, err)
			return 3
		}
	}

	_, _ = fmt.Fprintf(stdout, "Indexed %d words → %s\n", count, outputPath)
	return 0
}

// Run executes the CLI with the given OS primitives and returns an exit code.
func Run(args []string, stdin *os.File, stdout *os.File, stderr *os.File, getenv func(string) string) int {
	// Dispatch "index" subcommand before any main flag parsing.
	if len(args) > 0 && args[0] == "index" {
		return runIndex(args[1:], stdout, stderr)
	}

	cfg, err := parseConfig(args, getenv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %s\n", err.Error())
		return 2
	}

	if cfg.Help {
		_, _ = fmt.Fprintln(stdout, helpText)
		return 0
	}

	if cfg.Version {
		writeVersion(stdout, cfg.Output)
		return 0
	}

	useColor := UseColor(cfg.NoColor, getenv("NO_COLOR"), stdout)

	var formatter Formatter
	if cfg.FindSet {
		formatter = NewFormatterWithLookup(cfg.Output, useColor, cfg.Atbash, cfg.FindValue, cfg.Mispar)
	} else {
		formatter = NewFormatter(cfg.Output, useColor, cfg.Atbash)
	}

	if cfg.FindSet {
		return runFind(cfg, formatter, stdout, stderr, getenv)
	}

	if len(cfg.Args) > 0 {
		return computeArgs(cfg, formatter, stdout, stderr)
	}

	// No positional args: check whether stdin is interactive.
	if IsTerminal(stdin) {
		_, _ = fmt.Fprint(stderr, "Usage: gematria [OPTIONS] [INPUT...] — try 'gematria --help'\n")
		return 0
	}

	// Stdin is a pipe or redirect — process as batch.
	scanner := bufio.NewScanner(stdin)
	compute := computeFunc(cfg)
	return processBatch(scanner, compute, formatter, stdout, stderr, cfg.FailEarly)
}

// computeFunc returns the per-input compute closure for the active config.
// When --transliterate is set, the closure dispatches to ComputeTransliterated;
// otherwise to Compute. The same closure is shared by positional-arg and
// stdin-batch dispatch so behaviour is identical across both paths.
func computeFunc(cfg Config) func(string) (gematria.Result, error) {
	if cfg.Transliterate {
		scheme := gematria.Scheme(cfg.Scheme)
		return func(input string) (gematria.Result, error) {
			return gematria.ComputeTransliterated(input, cfg.Mispar, scheme)
		}
	}
	return func(input string) (gematria.Result, error) {
		return gematria.Compute(input, cfg.Mispar)
	}
}

// computeArgs computes gematria for each positional argument.
// On the first error it writes to stderr and returns the appropriate exit code
// without processing further arguments (no partial output on error).
func computeArgs(cfg Config, formatter Formatter, stdout, stderr *os.File) int {
	compute := computeFunc(cfg)
	for _, arg := range cfg.Args {
		result, err := compute(arg)
		if err != nil {
			_, _ = fmt.Fprint(stderr, formatter.FormatError(err))
			return exitCodeForComputeError(err)
		}
		_, _ = fmt.Fprint(stdout, formatter.FormatResult(result))
	}
	return 0
}

// exitCodeForComputeError maps a compute error to its exit code.
// Misuse-class errors (invalid system, invalid scheme) → 2; all other input
// errors (invalid char, unknown name, unknown word) → 1.
func exitCodeForComputeError(err error) int {
	var ise *gematria.InvalidSystemError
	var iscse *gematria.InvalidSchemeError
	if errors.As(err, &ise) || errors.As(err, &iscse) {
		return 2
	}
	return 1
}

// openWordSource selects and constructs the appropriate WordSource backend for path.
// Detection order (first match wins):
//  1. format is non-empty → use the specified backend explicitly
//  2. path begins with "http://" or "https://" → remote backend
//  3. path ends with ".db" → SQLite backend
//  4. path ends with ".idx" → index file backend
//  5. companion file path+".idx" exists → index file backend
//  6. default → in-memory backend (ParseWordList)
//
// Returns (source, closer, error). closer is non-nil only for backends that
// hold an open resource (SQLite DB, index file). The caller must call
// closer.Close() if closer != nil.
func openWordSource(path, format string, getenv func(string) string) (gematria.WordSource, io.Closer, error) {
	if format != "" {
		return openWordSourceByFormat(path, format, getenv)
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return openRemoteWordSource(path, getenv)
	}
	if strings.HasSuffix(path, ".db") {
		return openSQLiteWordSource(path)
	}
	if strings.HasSuffix(path, ".idx") {
		return openIndexWordSource(path)
	}
	// Check for companion .idx file.
	if _, err := os.Stat(path + ".idx"); err == nil {
		return openIndexWordSource(path + ".idx")
	}
	return openMemoryWordSource(path)
}

// openWordSourceByFormat constructs a backend by explicit format name.
func openWordSourceByFormat(path, format string, getenv func(string) string) (gematria.WordSource, io.Closer, error) {
	switch format {
	case "sqlite":
		return openSQLiteWordSource(path)
	case "index":
		return openIndexWordSource(path)
	case "remote":
		return openRemoteWordSource(path, getenv)
	default: // "memory"
		return openMemoryWordSource(path)
	}
}

func openSQLiteWordSource(path string) (gematria.WordSource, io.Closer, error) {
	src, err := gematria.OpenSQLiteWordSource(path)
	if err != nil {
		return nil, nil, err
	}
	return src, src.(io.Closer), nil
}

func openIndexWordSource(path string) (gematria.WordSource, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	src, err := gematria.NewIndexWordSource(f)
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return src, f, nil
}

func openRemoteWordSource(path string, getenv func(string) string) (gematria.WordSource, io.Closer, error) {
	var opts []gematria.RemoteOption
	if token := getenv("GEMATRIA_WORDLIST_TOKEN"); token != "" {
		opts = append(opts, gematria.WithAuthToken(token))
	}
	src, err := gematria.NewRemoteWordSource(path, opts...)
	if err != nil {
		return nil, nil, err
	}
	return src, nil, nil
}

func openMemoryWordSource(path string) (gematria.WordSource, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	src, err := gematria.ParseWordList(f)
	_ = f.Close()
	if err != nil {
		return nil, nil, err
	}
	return src, nil, nil
}

// runFind executes the reverse lookup branch when --find is active.
// It opens the word list via openWordSource, calls FindByValue, and renders
// results to stdout. Errors are written to stderr with appropriate exit codes.
func runFind(cfg Config, formatter Formatter, stdout, stderr *os.File, getenv func(string) string) int {
	if cfg.Wordlist == "" {
		_, _ = fmt.Fprint(stderr, formatter.FormatError(
			errors.New("--find requires a word list: use --wordlist <path> or set GEMATRIA_WORDLIST")))
		return 2
	}

	source, closer, err := openWordSource(cfg.Wordlist, cfg.WordlistFormat, getenv)
	if err != nil {
		_, _ = fmt.Fprint(stderr, formatter.FormatError(
			fmt.Errorf("cannot open word list %q: %w", cfg.Wordlist, err)))
		return 3
	}
	if closer != nil {
		defer func() { _ = closer.Close() }()
	}

	words, hasMore, err := gematria.FindByValue(cfg.FindValue, source, cfg.Mispar, cfg.Limit)
	if err != nil {
		_, _ = fmt.Fprint(stderr, formatter.FormatError(err))
		return 1
	}

	_, _ = fmt.Fprint(stdout, formatter.FormatLookup(words, hasMore))
	return 0
}

// writeVersion writes the version string to stdout in the appropriate format.
//
// The human form keeps the leading "v" (matches the git tag and the GitHub
// release page). The JSON form strips the "v" so machine-readable consumers
// see a value that matches the SemVer 2.0.0 spec (which forbids the prefix).
func writeVersion(stdout *os.File, output string) {
	if output == "json" {
		type versionJSON struct {
			Version string `json:"version"`
		}
		jsonVer := strings.TrimPrefix(cliVersion, "v")
		b, _ := json.Marshal(versionJSON{Version: jsonVer})
		_, _ = fmt.Fprintln(stdout, string(b))
	} else {
		_, _ = fmt.Fprintf(stdout, "gematria %s\n", cliVersion)
	}
}
