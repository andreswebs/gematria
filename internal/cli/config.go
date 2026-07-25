package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	gematria "github.com/andreswebs/gematria"
	"github.com/spf13/pflag"
)

// Config holds the resolved CLI configuration for a single invocation.
type Config struct {
	Mispar         gematria.System // selected gematria system (default: hechrachi)
	Output         string          // selected output format (default: "line")
	NoColor        bool            // disable ANSI color output
	Atbash         bool            // display Atbash substitution alongside normal output
	FailEarly      bool            // stop stdin batch on first error
	Help           bool            // print help and exit
	Version        bool            // print version and exit
	Args           []string        // remaining positional arguments
	FindValue      int             // numeric value to look up (only valid when FindSet=true)
	FindSet        bool            // true when --find was explicitly provided
	Wordlist       string          // resolved word list path (empty means none provided)
	Limit          int             // maximum results (always >= 1 after resolution)
	WordlistFormat string          // explicit backend override (sqlite|index|memory); empty = auto-detect
	Transliterate  bool            // interpret Latin input as Hebrew words (per --scheme)
	Scheme         string          // transliteration scheme (academic|israeli); default "academic" when Transliterate=true
	Index          bool            // build a pre-computed index from --wordlist
	IndexOutput    string          // explicit output file path for index (bypasses env var resolution)
	IndexFormat    string          // index format: sqlite|index (default: sqlite)
}

var validSystems = []string{"hechrachi", "gadol", "siduri", "atbash"}
var validOutputs = []string{"line", "value", "card", "json"}
var validWordlistFormats = []string{"sqlite", "index", "memory"}
var validSchemes = []string{"academic", "israeli"}
var validIndexFormats = []string{"sqlite", "index"}

// cliFlags holds the raw flag values bound during registration, before
// environment-variable and default resolution.
type cliFlags struct {
	mispar         string
	output         string
	noColor        bool
	atbash         bool
	failEarly      bool
	help           bool
	version        bool
	findValue      int
	wordlist       string
	limit          int
	wordlistFormat string
	transliterate  bool
	scheme         string
	index          bool
	indexOutput    string
	indexFormat    string
}

// registerFlags binds every CLI flag on fs to the fields of v. It is the
// single source of truth for flag registration; tests enumerate flags
// through it to guard against drift from the hand-written help text.
func registerFlags(fs *pflag.FlagSet, v *cliFlags) {
	fs.StringVarP(&v.mispar, "mispar", "m", "", "gematria system (hechrachi, gadol, siduri, atbash)")
	fs.StringVarP(&v.output, "output", "o", "", "output format (line, value, card, json)")
	fs.BoolVar(&v.noColor, "no-color", false, "disable color output")
	fs.BoolVar(&v.atbash, "atbash", false, "show Atbash letter substitutions in output")
	fs.BoolVar(&v.failEarly, "fail-early", false, "stop on first stdin error")
	fs.BoolVarP(&v.help, "help", "h", false, "show this help message")
	fs.BoolVar(&v.version, "version", false, "print version and exit")
	fs.IntVar(&v.findValue, "find", 0, "find words whose gematria value equals N")
	fs.StringVar(&v.wordlist, "wordlist", "", "path to word list file (or set GEMATRIA_WORDLIST)")
	fs.IntVarP(&v.limit, "limit", "l", 0, "maximum results to return (default 20)")
	fs.StringVar(&v.wordlistFormat, "wordlist-format", "", "backend override: sqlite|index|memory")
	fs.BoolVarP(&v.transliterate, "transliterate", "t", false, "interpret Latin input as Hebrew words (per --scheme)")
	fs.StringVar(&v.scheme, "scheme", "", "transliteration scheme (academic, israeli)")
	fs.BoolVar(&v.index, "index", false, "build a pre-computed index from --wordlist")
	fs.StringVar(&v.indexOutput, "index-output", "", "output file path for index")
	fs.StringVar(&v.indexFormat, "index-format", "sqlite", "index format: sqlite|index")
}

// parseConfig resolves flags and environment variables into a Config.
// Precedence: explicit flag > environment variable > built-in default.
// Invalid values return a typed error: *usageError for flag-level misuse
// (exit 64) and *configError for a bad environment-variable value (exit 78).
// Run classifies these by type, not by matching the message.
func parseConfig(args []string, getenv func(string) string) (Config, error) {
	fs := pflag.NewFlagSet("gematria", pflag.ContinueOnError)
	// Suppress pflag's own error/usage output; Run() handles all user-facing messages.
	fs.SetOutput(io.Discard)

	var v cliFlags
	registerFlags(fs, &v)

	if err := fs.Parse(args); err != nil {
		// pflag surface errors (unknown flag, bad numeric value) are usage errors.
		return Config{}, &usageError{err}
	}

	findSet := fs.Changed("find")

	// Resolve --output: flag > GEMATRIA_OUTPUT > "line"
	output := v.output
	outputFromFlag := fs.Changed("output")
	if output == "" {
		output = getenv("GEMATRIA_OUTPUT")
	}
	if output == "" {
		output = "line"
	}
	if !contains(validOutputs, output) {
		// Classify by provenance: an invalid flag value is usage (64), an
		// invalid GEMATRIA_OUTPUT value is configuration (78).
		if outputFromFlag {
			return Config{}, &usageError{fmt.Errorf("invalid value %q for --output\nvalid values: %s", output, strings.Join(validOutputs, ", "))}
		}
		return Config{}, &configError{fmt.Errorf("invalid value %q for GEMATRIA_OUTPUT\nvalid values: %s", output, strings.Join(validOutputs, ", "))}
	}

	// Resolve --mispar: flag > GEMATRIA_MISPAR > "hechrachi"
	// Only validate the flag value eagerly; GEMATRIA_MISPAR is validated lazily
	// when compute is actually invoked (an invalid env var should not block
	// unrelated operations such as reverse lookup).
	mispar := v.mispar
	if fs.Changed("mispar") && !contains(validSystems, mispar) {
		return Config{}, &usageError{fmt.Errorf("invalid value %q for --mispar\nvalid values: %s", mispar, strings.Join(validSystems, ", "))}
	}
	if mispar == "" {
		mispar = getenv("GEMATRIA_MISPAR")
	}
	if mispar == "" {
		mispar = "hechrachi"
	}

	// Resolve --wordlist: flag > GEMATRIA_WORDLIST (lazy: path not validated here)
	wordlist := v.wordlist
	if !fs.Changed("wordlist") {
		wordlist = getenv("GEMATRIA_WORDLIST")
	}

	// Validate --wordlist-format if explicitly provided.
	if v.wordlistFormat != "" && !contains(validWordlistFormats, v.wordlistFormat) {
		return Config{}, &usageError{fmt.Errorf("invalid value %q for --wordlist-format\nvalid values: %s", v.wordlistFormat, strings.Join(validWordlistFormats, ", "))}
	}

	// Validate --index-format when explicitly set.
	if fs.Changed("index-format") && !contains(validIndexFormats, v.indexFormat) {
		return Config{}, &usageError{fmt.Errorf("invalid value %q for --index-format\nvalid values: %s", v.indexFormat, strings.Join(validIndexFormats, ", "))}
	}

	// --index conflict checks (all require wordlist to already be resolved above).
	if v.index && findSet {
		return Config{}, &usageError{fmt.Errorf("--index and --find are mutually exclusive")}
	}
	if v.index && v.transliterate {
		return Config{}, &usageError{fmt.Errorf("--index and --transliterate are mutually exclusive")}
	}
	if fs.Changed("index-output") && !v.index {
		return Config{}, &usageError{fmt.Errorf("--index-output requires --index")}
	}
	if fs.Changed("index-format") && !v.index {
		return Config{}, &usageError{fmt.Errorf("--index-format requires --index")}
	}
	if v.index && len(fs.Args()) > 0 {
		return Config{}, &usageError{fmt.Errorf("--index does not accept positional arguments")}
	}
	if v.index && wordlist == "" {
		return Config{}, &usageError{fmt.Errorf("--index requires --wordlist or GEMATRIA_WORDLIST")}
	}

	// Resolve --limit: flag > GEMATRIA_LIMIT > DefaultLookupLimit
	// GEMATRIA_LIMIT is validated lazily: only parsed and validated when --find is active.
	limit := v.limit
	if !fs.Changed("limit") && findSet {
		if envLimit := getenv("GEMATRIA_LIMIT"); envLimit != "" {
			parsed, err := strconv.Atoi(envLimit)
			if err != nil || parsed < 1 {
				return Config{}, &configError{fmt.Errorf("GEMATRIA_LIMIT must be a positive integer, got: %q", envLimit)}
			}
			limit = parsed
		}
	}
	if limit <= 0 {
		limit = gematria.DefaultLookupLimit
	}

	// Resolve --scheme: flag > GEMATRIA_SCHEME > "academic" (when -t active)
	// Eager validation when the flag was explicitly set.
	scheme := v.scheme
	if fs.Changed("scheme") && !contains(validSchemes, scheme) {
		return Config{}, &usageError{fmt.Errorf("invalid value %q for --scheme\nvalid values: %s", scheme, strings.Join(validSchemes, ", "))}
	}
	// Fall back to env var when flag not provided.
	if scheme == "" {
		scheme = getenv("GEMATRIA_SCHEME")
	}
	// Lazy validation: only check env var value when -t is active.
	if v.transliterate && scheme != "" && !contains(validSchemes, scheme) {
		return Config{}, &configError{fmt.Errorf("invalid value %q for GEMATRIA_SCHEME\nvalid values: %s", scheme, strings.Join(validSchemes, ", "))}
	}
	// Apply default only when transliteration is active.
	if v.transliterate && scheme == "" {
		scheme = "academic"
	}

	return Config{
		Mispar:         gematria.System(mispar),
		Output:         output,
		NoColor:        v.noColor,
		Atbash:         v.atbash,
		FailEarly:      v.failEarly,
		Help:           v.help,
		Version:        v.version,
		Args:           fs.Args(),
		FindValue:      v.findValue,
		FindSet:        findSet,
		Wordlist:       wordlist,
		Limit:          limit,
		WordlistFormat: v.wordlistFormat,
		Transliterate:  v.transliterate,
		Scheme:         scheme,
		Index:          v.index,
		IndexOutput:    v.indexOutput,
		IndexFormat:    v.indexFormat,
	}, nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
