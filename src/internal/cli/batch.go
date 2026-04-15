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
// Return codes:
//   - 0: all lines succeeded (or no lines)
//   - 1: all lines failed (or failEarly stopped on InvalidCharError/UnknownNameError)
//   - 2: failEarly stopped on InvalidSystemError
//   - 4: partial success (some lines succeeded, some failed)
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
		return 4
	case errorCount > 0:
		return 1
	default:
		return 0
	}
}

// exitCodeForBatchError maps a compute error to the appropriate exit code for
// failEarly mode. Misuse-class errors (invalid system, invalid scheme) → 2;
// all other input errors (invalid char, unknown name, unknown word) → 1.
func exitCodeForBatchError(err error) int {
	var ise *gematria.InvalidSystemError
	var iscse *gematria.InvalidSchemeError
	if errors.As(err, &ise) || errors.As(err, &iscse) {
		return 2
	}
	return 1
}
