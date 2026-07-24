package cli

import (
	"fmt"
	"strings"

	gematria "github.com/andreswebs/gematria"
)

// cardFormatter renders results as an aligned per-letter table with a total
// line. Intended for human study and reference.
type cardFormatter struct {
	useColor   bool
	showAtbash bool
	findValue  int
	findSystem gematria.System
}

// Column widths (visible characters).
const (
	cardColLetter = 6  // Hebrew char + surrounding spaces; RTL/LTR are zero-width
	cardColName   = 12 // "Tsade Sofit" is longest at 11
	cardColValue  = 5  // max value 900 in Gadol, right-justified
)

func (f *cardFormatter) FormatResult(r gematria.Result) string {
	var sb strings.Builder

	// Header
	if f.showAtbash {
		fmt.Fprintf(&sb, "%-*s  %-*s  %*s  %-s  %-s\n",
			cardColLetter, "Letter",
			cardColName, "Name",
			cardColValue, "Value",
			"Atbash",
			"Meaning",
		)
	} else {
		fmt.Fprintf(&sb, "%-*s  %-*s  %*s  %-s\n",
			cardColLetter, "Letter",
			cardColName, "Name",
			cardColValue, "Value",
			"Meaning",
		)
	}

	// Rows
	for _, lr := range r.Letters {
		heb := wrapHebrew(string(lr.Char))
		name := lr.Name
		val := fmt.Sprintf("%d", lr.Value)
		meaning := lr.Meaning

		if f.useColor {
			heb = applyColor(ansiBold, heb)
			val = applyColor(ansiGreen, val)
		}

		if f.showAtbash {
			atbashChar := wrapHebrew(string(gematria.AtbashSubstitute(lr.Char)))
			if f.useColor {
				atbashChar = applyColor(ansiBold, atbashChar)
			}
			fmt.Fprintf(&sb, "%-*s  %-*s  %*s  %-6s  %-s\n",
				cardColLetter, heb,
				cardColName, name,
				cardColValue, val,
				atbashChar,
				meaning,
			)
		} else {
			fmt.Fprintf(&sb, "%-*s  %-*s  %*s  %-s\n",
				cardColLetter, heb,
				cardColName, name,
				cardColValue, val,
				meaning,
			)
		}
	}

	// Separator
	sepWidth := cardColLetter + 2 + cardColName + 2 + cardColValue + 2 + 12
	if f.showAtbash {
		sepWidth += 6 + 2
	}
	sb.WriteString(strings.Repeat("\u2500", sepWidth) + "\n")

	// Total line
	total := fmt.Sprintf("%d", r.Total)
	if f.useColor {
		total = applyColor(ansiGreen, total)
	}
	fmt.Fprintf(&sb, "Total: %s  [%s]\n", total, r.System)

	// Scheme line — only present when transliteration was used.
	if r.Scheme != "" {
		fmt.Fprintf(&sb, "Scheme: %s\n", r.Scheme)
	}

	return sb.String()
}

func (f *cardFormatter) FormatLookup(words []gematria.Word, hasMore bool) string {
	if len(words) == 0 {
		return "(no results)\n"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Reverse Lookup: %d\n", f.findValue)
	fmt.Fprintf(&sb, "System: %s\n", f.findSystem)
	for i, w := range words {
		sb.WriteString("\n")
		heb := wrapHebrew(w.Hebrew)
		if f.useColor {
			heb = applyColor(ansiBold, heb)
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, heb)
		if w.Transliteration != "" {
			fmt.Fprintf(&sb, "   Transliteration: %s\n", w.Transliteration)
		}
		if w.Meaning != "" {
			fmt.Fprintf(&sb, "   Meaning: %s\n", w.Meaning)
		}
	}
	if hasMore {
		sb.WriteString("\n(more results available \u2014 increase --limit to see them)\n")
	}
	return sb.String()
}

func (f *cardFormatter) FormatError(err error) string {
	return formatErrorPlain(err, f.useColor)
}
