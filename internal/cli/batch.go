package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	gematria "github.com/andreswebs/gematria"
)

// batchLineError wraps a compute error with its 1-based source line number.
// Formatters receive this as the err argument to FormatError so they can
// include the line number in their output.
type batchLineError struct {
	cause   error
	lineNum int
}

func (e *batchLineError) Error() string {
	return fmt.Sprintf("line %d: %s", e.lineNum, e.cause.Error())
}

func (e *batchLineError) Unwrap() error {
	return e.cause
}

// processBatch reads lines from scanner, applies compute to each, and writes
// results to stdout or errors (with line number) to stderr.
//
// Return codes (per docs/adr/0001-exit-code-taxonomy.md):
//   - 0:  all lines succeeded (or no lines)
//   - 1:  partial success (some lines succeeded, some failed) — recoverable result
//   - 65: all lines failed — data error (or failEarly stopped on invalid char/name/word)
//   - 78: failEarly stopped on invalid system/scheme from an env var
func processBatch(
	scanner *bufio.Scanner,
	compute func(string) (gematria.Result, error),
	formatter Formatter,
	stdout, stderr *os.File,
	failEarly bool,
) int {
	lineNum := 0
	successCount := 0
	errorCount := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		result, err := compute(line)
		if err != nil {
			errorCount++
			wrapped := &batchLineError{cause: err, lineNum: lineNum}
			_, _ = fmt.Fprint(stderr, formatter.FormatError(wrapped))
			if failEarly {
				return exitCodeForBatchError(err)
			}
			continue
		}
		successCount++
		_, _ = fmt.Fprint(stdout, formatter.FormatResult(result))
	}

	switch {
	case errorCount > 0 && successCount > 0:
		return exitPartialBatch
	case errorCount > 0:
		return exitDataErr
	default:
		return exitOK
	}
}

// exitCodeForBatchError maps a compute error to the appropriate exit code for
// failEarly mode. An invalid system/scheme can only reach compute time via a
// lazily validated env var (flag values are validated eagerly at parse time),
// so it is a configuration error (78); all other input errors (invalid char,
// unknown name, unknown word) are data errors (65).
func exitCodeForBatchError(err error) int {
	var ise *gematria.InvalidSystemError
	var iscse *gematria.InvalidSchemeError
	if errors.As(err, &ise) || errors.As(err, &iscse) {
		return exitConfigErr
	}
	return exitDataErr
}
