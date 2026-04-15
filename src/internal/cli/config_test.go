package cli

import (
	"strings"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

func noenv(string) string { return "" }

func envWith(kv map[string]string) func(string) string {
	return func(key string) string { return kv[key] }
}

func TestParseConfig_Defaults(t *testing.T) {
	cfg, err := parseConfig([]string{}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mispar != gematria.Hechrachi {
		t.Errorf("Mispar: got %q, want %q", cfg.Mispar, gematria.Hechrachi)
	}
	if cfg.Output != "line" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "line")
	}
	if cfg.NoColor {
		t.Error("NoColor: want false, got true")
	}
	if cfg.Atbash {
		t.Error("Atbash: want false, got true")
	}
	if cfg.FailEarly {
		t.Error("FailEarly: want false, got true")
	}
	if cfg.Version {
		t.Error("Version: want false, got true")
	}
	if len(cfg.Args) != 0 {
		t.Errorf("Args: want empty, got %v", cfg.Args)
	}
}

func TestParseConfig_MisparPrecedence(t *testing.T) {
	// Flag overrides env var and default.
	cfg, err := parseConfig([]string{"--mispar", "gadol"}, envWith(map[string]string{
		"GEMATRIA_MISPAR": "siduri",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mispar != gematria.Gadol {
		t.Errorf("Mispar: got %q, want %q", cfg.Mispar, gematria.Gadol)
	}

	// Env var used when flag not set.
	cfg, err = parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_MISPAR": "siduri",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mispar != gematria.Siduri {
		t.Errorf("Mispar: got %q, want %q", cfg.Mispar, gematria.Siduri)
	}
}

func TestParseConfig_OutputPrecedence(t *testing.T) {
	// Flag overrides env var and default.
	cfg, err := parseConfig([]string{"--output", "json"}, envWith(map[string]string{
		"GEMATRIA_OUTPUT": "card",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "json")
	}

	// Env var used when flag not set.
	cfg, err = parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_OUTPUT": "value",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != "value" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "value")
	}
}

func TestParseConfig_InvalidMispar(t *testing.T) {
	_, err := parseConfig([]string{"--mispar", "standard"}, noenv)
	if err == nil {
		t.Fatal("expected error for invalid --mispar, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "standard") {
		t.Errorf("error should mention invalid value 'standard': %q", msg)
	}
	if !strings.Contains(msg, "--mispar") {
		t.Errorf("error should mention flag name: %q", msg)
	}
	for _, v := range []string{"hechrachi", "gadol", "siduri", "atbash"} {
		if !strings.Contains(msg, v) {
			t.Errorf("error should list valid value %q: %q", v, msg)
		}
	}
}

func TestParseConfig_InvalidOutput(t *testing.T) {
	_, err := parseConfig([]string{"--output", "table"}, noenv)
	if err == nil {
		t.Fatal("expected error for invalid --output, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "table") {
		t.Errorf("error should mention invalid value 'table': %q", msg)
	}
	if !strings.Contains(msg, "--output") {
		t.Errorf("error should mention flag name: %q", msg)
	}
	for _, v := range []string{"line", "value", "card", "json"} {
		if !strings.Contains(msg, v) {
			t.Errorf("error should list valid value %q: %q", v, msg)
		}
	}
}

func TestParseConfig_BoolFlags(t *testing.T) {
	cfg, err := parseConfig([]string{"--no-color", "--atbash", "--fail-early", "--version"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.NoColor {
		t.Error("NoColor: want true, got false")
	}
	if !cfg.Atbash {
		t.Error("Atbash: want true, got false")
	}
	if !cfg.FailEarly {
		t.Error("FailEarly: want true, got false")
	}
	if !cfg.Version {
		t.Error("Version: want true, got false")
	}
}

func TestParseConfig_PositionalArgs(t *testing.T) {
	cfg, err := parseConfig([]string{"--mispar", "gadol", "שלום", "shalom"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"שלום", "shalom"}
	if len(cfg.Args) != len(want) {
		t.Fatalf("Args length: got %d, want %d: %v", len(cfg.Args), len(want), cfg.Args)
	}
	for i, a := range cfg.Args {
		if a != want[i] {
			t.Errorf("Args[%d]: got %q, want %q", i, a, want[i])
		}
	}
}

func TestParseConfig_InvalidMisparEnvIsLazy(t *testing.T) {
	// GEMATRIA_MISPAR with an invalid value must NOT cause an error in parseConfig;
	// validation is deferred until compute is invoked.
	cfg, err := parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_MISPAR": "invalid-system",
	}))
	if err != nil {
		t.Fatalf("GEMATRIA_MISPAR invalid value should not error in parseConfig, got: %v", err)
	}
	if string(cfg.Mispar) != "invalid-system" {
		t.Errorf("Mispar: got %q, want %q", cfg.Mispar, "invalid-system")
	}
}

func TestParseConfig_InvalidOutputEnvIsEager(t *testing.T) {
	// GEMATRIA_OUTPUT with an invalid value MUST cause an error in parseConfig;
	// it is needed to determine how errors themselves are rendered.
	_, err := parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_OUTPUT": "csv",
	}))
	if err == nil {
		t.Fatal("expected error for invalid GEMATRIA_OUTPUT, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "csv") {
		t.Errorf("error should mention invalid value: %q", msg)
	}
	for _, v := range []string{"line", "value", "card", "json"} {
		if !strings.Contains(msg, v) {
			t.Errorf("error should list valid value %q: %q", v, msg)
		}
	}
}

func TestParseConfig_ShortFlags(t *testing.T) {
	cfg, err := parseConfig([]string{"-m", "atbash", "-o", "json"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mispar != gematria.Atbash {
		t.Errorf("Mispar: got %q, want %q", cfg.Mispar, gematria.Atbash)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "json")
	}
}

// --- --find flag ---

func TestParseConfig_FindFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.FindSet {
		t.Error("FindSet: want true, got false")
	}
	if cfg.FindValue != 376 {
		t.Errorf("FindValue: got %d, want 376", cfg.FindValue)
	}
}

func TestParseConfig_FindNotSet_Defaults(t *testing.T) {
	cfg, err := parseConfig([]string{}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FindSet {
		t.Error("FindSet: want false, got true")
	}
	if cfg.FindValue != 0 {
		t.Errorf("FindValue: got %d, want 0", cfg.FindValue)
	}
}

// --- --wordlist flag ---

func TestParseConfig_WordlistFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--wordlist", "/path/to/words.txt"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Wordlist != "/path/to/words.txt" {
		t.Errorf("Wordlist: got %q, want %q", cfg.Wordlist, "/path/to/words.txt")
	}
}

func TestParseConfig_WordlistEnvVar(t *testing.T) {
	cfg, err := parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_WORDLIST": "/env/words.txt",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Wordlist != "/env/words.txt" {
		t.Errorf("Wordlist: got %q, want %q", cfg.Wordlist, "/env/words.txt")
	}
}

func TestParseConfig_WordlistFlagOverridesEnv(t *testing.T) {
	cfg, err := parseConfig([]string{"--wordlist", "/flag/words.txt"}, envWith(map[string]string{
		"GEMATRIA_WORDLIST": "/env/words.txt",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Wordlist != "/flag/words.txt" {
		t.Errorf("Wordlist: got %q, want %q", cfg.Wordlist, "/flag/words.txt")
	}
}

// --- --limit / -l flag ---

func TestParseConfig_LimitFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376", "--limit", "5"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 5 {
		t.Errorf("Limit: got %d, want 5", cfg.Limit)
	}
}

func TestParseConfig_LimitShortFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376", "-l", "10"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 10 {
		t.Errorf("Limit: got %d, want 10", cfg.Limit)
	}
}

func TestParseConfig_LimitDefaultIs20(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != gematria.DefaultLookupLimit {
		t.Errorf("Limit: got %d, want %d", cfg.Limit, gematria.DefaultLookupLimit)
	}
}

func TestParseConfig_LimitEnvVar(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376"}, envWith(map[string]string{
		"GEMATRIA_LIMIT": "50",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 50 {
		t.Errorf("Limit: got %d, want 50", cfg.Limit)
	}
}

func TestParseConfig_LimitFlagOverridesEnv(t *testing.T) {
	cfg, err := parseConfig([]string{"--find", "376", "--limit", "5"}, envWith(map[string]string{
		"GEMATRIA_LIMIT": "50",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != 5 {
		t.Errorf("Limit: got %d, want 5", cfg.Limit)
	}
}

func TestParseConfig_LimitEnvInvalidWhenFindSet(t *testing.T) {
	_, err := parseConfig([]string{"--find", "376"}, envWith(map[string]string{
		"GEMATRIA_LIMIT": "notanumber",
	}))
	if err == nil {
		t.Fatal("expected error for invalid GEMATRIA_LIMIT when --find is set, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "GEMATRIA_LIMIT") {
		t.Errorf("error should mention GEMATRIA_LIMIT: %q", msg)
	}
}

func TestParseConfig_LimitEnvNotValidatedWhenFindNotSet(t *testing.T) {
	// GEMATRIA_LIMIT is lazy: invalid value must NOT error when --find is not set.
	cfg, err := parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_LIMIT": "bad",
	}))
	if err != nil {
		t.Fatalf("GEMATRIA_LIMIT invalid value should not error without --find, got: %v", err)
	}
	// Limit should default to DefaultLookupLimit when not parsed from env.
	if cfg.Limit != gematria.DefaultLookupLimit {
		t.Errorf("Limit: got %d, want %d", cfg.Limit, gematria.DefaultLookupLimit)
	}
}

func TestParseConfig_LimitZeroReplacedByDefault(t *testing.T) {
	// Explicit --limit 0 should be replaced by DefaultLookupLimit.
	cfg, err := parseConfig([]string{"--find", "376", "--limit", "0"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Limit != gematria.DefaultLookupLimit {
		t.Errorf("Limit: got %d, want %d", cfg.Limit, gematria.DefaultLookupLimit)
	}
}

// --- --transliterate / -t flag and --scheme / GEMATRIA_SCHEME ---

func TestParseConfig_TransliterateShortFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"-t"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Transliterate {
		t.Error("Transliterate: want true, got false")
	}
}

func TestParseConfig_TransliterateLongFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--transliterate"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Transliterate {
		t.Error("Transliterate: want true, got false")
	}
}

func TestParseConfig_TransliterateDefault(t *testing.T) {
	cfg, err := parseConfig([]string{}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transliterate {
		t.Error("Transliterate: want false, got true")
	}
}

func TestParseConfig_SchemeDefaultWhenTransliterateSet(t *testing.T) {
	cfg, err := parseConfig([]string{"-t"}, noenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scheme != "academic" {
		t.Errorf("Scheme: got %q, want %q", cfg.Scheme, "academic")
	}
}

func TestParseConfig_SchemeFlag(t *testing.T) {
	for _, scheme := range []string{"academic", "israeli"} {
		cfg, err := parseConfig([]string{"-t", "--scheme", scheme}, noenv)
		if err != nil {
			t.Fatalf("--scheme %s: unexpected error: %v", scheme, err)
		}
		if cfg.Scheme != scheme {
			t.Errorf("--scheme %s: got %q, want %q", scheme, cfg.Scheme, scheme)
		}
	}
}

func TestParseConfig_InvalidSchemeEager(t *testing.T) {
	// --scheme with an invalid value must error immediately, even without -t.
	_, err := parseConfig([]string{"--scheme", "bogus"}, noenv)
	if err == nil {
		t.Fatal("expected error for --scheme bogus, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bogus") {
		t.Errorf("error should mention invalid value 'bogus': %q", msg)
	}
	if !strings.Contains(msg, "--scheme") {
		t.Errorf("error should mention flag name: %q", msg)
	}
	for _, v := range []string{"academic", "israeli"} {
		if !strings.Contains(msg, v) {
			t.Errorf("error should list valid value %q: %q", v, msg)
		}
	}
}

func TestParseConfig_SchemeEnvVar(t *testing.T) {
	cfg, err := parseConfig([]string{"-t"}, envWith(map[string]string{
		"GEMATRIA_SCHEME": "israeli",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scheme != "israeli" {
		t.Errorf("Scheme: got %q, want %q", cfg.Scheme, "israeli")
	}
}

func TestParseConfig_SchemeFlagOverridesEnv(t *testing.T) {
	cfg, err := parseConfig([]string{"-t", "--scheme", "academic"}, envWith(map[string]string{
		"GEMATRIA_SCHEME": "israeli",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scheme != "academic" {
		t.Errorf("Scheme: got %q, want %q", cfg.Scheme, "academic")
	}
}

func TestParseConfig_SchemeEnvInvalidWithoutTransliterate(t *testing.T) {
	// GEMATRIA_SCHEME with invalid value must NOT error when -t is not active.
	cfg, err := parseConfig([]string{}, envWith(map[string]string{
		"GEMATRIA_SCHEME": "bogus",
	}))
	if err != nil {
		t.Fatalf("GEMATRIA_SCHEME invalid without -t should not error, got: %v", err)
	}
	// Scheme is stored as-is (not defaulted to academic since -t is off).
	if cfg.Scheme != "bogus" {
		t.Errorf("Scheme: got %q, want %q", cfg.Scheme, "bogus")
	}
}

func TestParseConfig_SchemeEnvInvalidWithTransliterate(t *testing.T) {
	// GEMATRIA_SCHEME with invalid value MUST error when -t is active.
	_, err := parseConfig([]string{"-t"}, envWith(map[string]string{
		"GEMATRIA_SCHEME": "bogus",
	}))
	if err == nil {
		t.Fatal("expected error for invalid GEMATRIA_SCHEME with -t, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bogus") {
		t.Errorf("error should mention invalid value: %q", msg)
	}
	if !strings.Contains(msg, "GEMATRIA_SCHEME") {
		t.Errorf("error should mention env var name: %q", msg)
	}
	for _, v := range []string{"academic", "israeli"} {
		if !strings.Contains(msg, v) {
			t.Errorf("error should list valid value %q: %q", v, msg)
		}
	}
}
