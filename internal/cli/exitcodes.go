package cli

// Exit codes follow the fleet taxonomy in docs/adr/0001-exit-code-taxonomy.md.
// 0 success; 1 recoverable result (partial batch); 64 usage; 65 data;
// 74 I/O; 78 configuration. This tool has no internal-error (70) path.
const (
	exitOK           = 0  // success, clean
	exitPartialBatch = 1  // recoverable result: some batch lines failed
	exitUsage        = 64 // EX_USAGE: the CLI surface was misused
	exitDataErr      = 65 // EX_DATAERR: input or payload data rejected
	exitIOErr        = 74 // EX_IOERR: filesystem or backend I/O failure
	exitConfigErr    = 78 // EX_CONFIG: invalid environment/config value
)

// ExitCodes declares every exit code this tool can produce,
// per docs/adr/0001-exit-code-taxonomy.md. It is the machine-readable
// registry required by the ADR; the conformance test in exitcodes_test.go
// asserts that every code the tool actually emits is a key here.
var ExitCodes = map[int]string{
	exitOK:           "success",
	exitPartialBatch: "partial batch success",
	exitUsage:        "usage error",
	exitDataErr:      "data error (invalid input, unknown word, malformed wordlist, all batch lines failed)",
	exitIOErr:        "I/O error (wordlist/index/backend)",
	exitConfigErr:    "configuration error (invalid environment variable value)",
}

// usageError marks a flag-level misuse: a bad flag value, mutually exclusive
// flags, or a required input missing for the requested flag. It maps to
// exitUsage (64).
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

// configError marks an invalid environment-variable value (or env/XDG
// resolution failure). It maps to exitConfigErr (78). Distinguishing it from
// usageError lets Run classify a bad value by its provenance without
// string-matching the message.
type configError struct{ err error }

func (e *configError) Error() string { return e.err.Error() }
func (e *configError) Unwrap() error { return e.err }
