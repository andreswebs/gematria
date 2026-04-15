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
	WordlistFormat string          // explicit backend override (sqlite|index|remote|memory); empty = auto-detect
	Transliterate  bool            // interpret Latin input as Hebrew words (per --scheme)
	Scheme         string          // transliteration scheme (academic|israeli); default "academic" when Transliterate=true
}

var validSystems = []string{"hechrachi", "gadol", "siduri", "atbash"}
var validOutputs = []string{"line", "value", "card", "json"}
var validWordlistFormats = []string{"sqlite", "index", "remote", "memory"}
var validSchemes = []string{"academic", "israeli"}

// parseConfig resolves flags and environment variables into a Config.
// Precedence: explicit flag > environment variable > built-in default.
// Returns a non-nil error (with exit-code-2 semantics) for invalid values.
func parseConfig(args []string, getenv func(string) string) (Config, error) {
	fs := pflag.NewFlagSet("gematria", pflag.ContinueOnError)
	// Suppress pflag's own error/usage output; Run() handles all user-facing messages.
	fs.SetOutput(io.Discard)

	var misparFlag string
	var outputFlag string
	var noColor bool
	var atbash bool
	var failEarly bool
	var help bool
	var version bool
	var findValue int
	var wordlist string
	var limit int
	var wordlistFormat string
	var transliterate bool
	var schemeFlag string

	fs.StringVarP(&misparFlag, "mispar", "m", "", "gematria system (hechrachi, gadol, siduri, atbash)")
	fs.StringVarP(&outputFlag, "output", "o", "", "output format (line, value, card, json)")
	fs.BoolVar(&noColor, "no-color", false, "disable color output")
	fs.BoolVar(&atbash, "atbash", false, "show Atbash letter substitutions in output")
	fs.BoolVar(&failEarly, "fail-early", false, "stop on first stdin error")
	fs.BoolVarP(&help, "help", "h", false, "show this help message")
	fs.BoolVar(&version, "version", false, "print version and exit")
	fs.IntVar(&findValue, "find", 0, "find words whose gematria value equals N")
	fs.StringVar(&wordlist, "wordlist", "", "path to word list file (or set GEMATRIA_WORDLIST)")
	fs.IntVarP(&limit, "limit", "l", 0, "maximum results to return (default 20)")
	fs.StringVar(&wordlistFormat, "wordlist-format", "", "backend override: sqlite|index|remote|memory")
	fs.BoolVarP(&transliterate, "transliterate", "t", false, "interpret Latin input as Hebrew words (per --scheme)")
	fs.StringVar(&schemeFlag, "scheme", "", "transliteration scheme (academic, israeli)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	findSet := fs.Changed("find")

	// Resolve --output: flag > GEMATRIA_OUTPUT > "line"
	output := outputFlag
	if output == "" {
		output = getenv("GEMATRIA_OUTPUT")
	}
	if output == "" {
		output = "line"
	}
	if output != "" && !contains(validOutputs, output) {
		return Config{}, fmt.Errorf("invalid value %q for --output\nvalid values: %s", output, strings.Join(validOutputs, ", "))
	}

	// Resolve --mispar: flag > GEMATRIA_MISPAR > "hechrachi"
	// Only validate the flag value eagerly; GEMATRIA_MISPAR is validated lazily
	// when compute is actually invoked (an invalid env var should not block
	// unrelated operations such as reverse lookup).
	mispar := misparFlag
	if fs.Changed("mispar") && !contains(validSystems, mispar) {
		return Config{}, fmt.Errorf("invalid value %q for --mispar\nvalid values: %s", mispar, strings.Join(validSystems, ", "))
	}
	if mispar == "" {
		mispar = getenv("GEMATRIA_MISPAR")
	}
	if mispar == "" {
		mispar = "hechrachi"
	}

	// Resolve --wordlist: flag > GEMATRIA_WORDLIST (lazy: path not validated here)
	if !fs.Changed("wordlist") {
		wordlist = getenv("GEMATRIA_WORDLIST")
	}

	// Validate --wordlist-format if explicitly provided.
	if wordlistFormat != "" && !contains(validWordlistFormats, wordlistFormat) {
		return Config{}, fmt.Errorf("invalid value %q for --wordlist-format\nvalid values: %s", wordlistFormat, strings.Join(validWordlistFormats, ", "))
	}

	// Resolve --limit: flag > GEMATRIA_LIMIT > DefaultLookupLimit
	// GEMATRIA_LIMIT is validated lazily: only parsed and validated when --find is active.
	if !fs.Changed("limit") && findSet {
		if envLimit := getenv("GEMATRIA_LIMIT"); envLimit != "" {
			parsed, err := strconv.Atoi(envLimit)
			if err != nil || parsed < 1 {
				return Config{}, fmt.Errorf("GEMATRIA_LIMIT must be a positive integer, got: %q", envLimit)
			}
			limit = parsed
		}
	}
	if limit <= 0 {
		limit = gematria.DefaultLookupLimit
	}

	// Resolve --scheme: flag > GEMATRIA_SCHEME > "academic" (when -t active)
	// Eager validation when the flag was explicitly set.
	scheme := schemeFlag
	if fs.Changed("scheme") && !contains(validSchemes, scheme) {
		return Config{}, fmt.Errorf("invalid value %q for --scheme\nvalid values: %s", scheme, strings.Join(validSchemes, ", "))
	}
	// Fall back to env var when flag not provided.
	if scheme == "" {
		scheme = getenv("GEMATRIA_SCHEME")
	}
	// Lazy validation: only check env var value when -t is active.
	if transliterate && scheme != "" && !contains(validSchemes, scheme) {
		return Config{}, fmt.Errorf("invalid value %q for GEMATRIA_SCHEME\nvalid values: %s", scheme, strings.Join(validSchemes, ", "))
	}
	// Apply default only when transliteration is active.
	if transliterate && scheme == "" {
		scheme = "academic"
	}

	return Config{
		Mispar:         gematria.System(mispar),
		Output:         output,
		NoColor:        noColor,
		Atbash:         atbash,
		FailEarly:      failEarly,
		Help:           help,
		Version:        version,
		Args:           fs.Args(),
		FindValue:      findValue,
		FindSet:        findSet,
		Wordlist:       wordlist,
		Limit:          limit,
		WordlistFormat: wordlistFormat,
		Transliterate:  transliterate,
		Scheme:         scheme,
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
