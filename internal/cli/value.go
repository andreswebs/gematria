package cli

import (
	"fmt"
	"strings"

	gematria "github.com/andreswebs/gematria"
)

// valueFormatter renders only the bare numeric total — no labels, no Hebrew,
// no color. Intended for scripting and agent pipelines.
type valueFormatter struct{}

func (f *valueFormatter) FormatResult(r gematria.Result) string {
	return fmt.Sprintf("%d\n", r.Total)
}

func (f *valueFormatter) FormatLookup(words []gematria.Word, _ bool) string {
	var sb strings.Builder
	for _, w := range words {
		sb.WriteString(w.Hebrew)
		sb.WriteRune('\n')
	}
	return sb.String()
}

func (f *valueFormatter) FormatError(err error) string {
	return fmt.Sprintf("Error: %s", err.Error())
}
