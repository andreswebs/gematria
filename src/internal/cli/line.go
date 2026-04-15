package cli

import (
	"fmt"
	"strings"

	gematria "github.com/andreswebs/gematria"
)

// lineFormatter renders results as a single human-readable line with
// Hebrew RTL marks and an optional per-letter breakdown.
type lineFormatter struct {
	useColor   bool
	showAtbash bool
}

func (f *lineFormatter) FormatResult(r gematria.Result) string {
	if len(r.Letters) == 1 {
		return f.formatSingleLetter(r.Letters[0])
	}
	return f.formatWord(r)
}

// formatSingleLetter produces: "Name (‏Char‎) = Value"
func (f *lineFormatter) formatSingleLetter(lr gematria.LetterResult) string {
	heb := wrapHebrew(string(lr.Char))
	val := fmt.Sprintf("%d", lr.Value)
	if f.useColor {
		heb = applyColor(ansiBold, heb)
		val = applyColor(ansiGreen, val)
	}
	line := fmt.Sprintf("%s (%s) = %s", lr.Name, heb, val)
	if f.showAtbash {
		mirror := wrapHebrew(string(gematria.AtbashSubstitute(lr.Char)))
		if f.useColor {
			mirror = applyColor(ansiBold, mirror)
		}
		line += " \u2192 " + mirror
	}
	return line + "\n"
}

// formatWord produces: "‏word‎ = Total (‏l1‎=v1 + ‏l2‎=v2 ...)"
func (f *lineFormatter) formatWord(r gematria.Result) string {
	word := wrapHebrew(r.Input)
	total := fmt.Sprintf("%d", r.Total)
	if f.useColor {
		word = applyColor(ansiBold, word)
		total = applyColor(ansiGreen, total)
	}

	parts := make([]string, len(r.Letters))
	for i, lr := range r.Letters {
		heb := wrapHebrew(string(lr.Char))
		val := fmt.Sprintf("%d", lr.Value)
		if f.useColor {
			heb = applyColor(ansiBold, heb)
			val = applyColor(ansiGreen, val)
		}
		parts[i] = heb + "=" + val
	}

	line := fmt.Sprintf("%s = %s (%s)", word, total, strings.Join(parts, " + "))

	if f.showAtbash {
		mirrors := ""
		for _, lr := range r.Letters {
			mirrors += string(gematria.AtbashSubstitute(lr.Char))
		}
		m := wrapHebrew(mirrors)
		if f.useColor {
			m = applyColor(ansiBold, m)
		}
		line += " \u2192 " + m
	}
	return line + "\n"
}

func (f *lineFormatter) FormatLookup(words []gematria.Word, hasMore bool) string {
	if len(words) == 0 {
		return "(no results)\n"
	}
	var sb strings.Builder
	for _, w := range words {
		heb := wrapHebrew(w.Hebrew)
		if f.useColor {
			heb = applyColor(ansiGreen, heb)
		}
		line := heb
		if w.Transliteration != "" {
			line += " (" + w.Transliteration + ")"
		}
		if w.Meaning != "" {
			line += " \u2014 " + w.Meaning
		}
		sb.WriteString(line + "\n")
	}
	if hasMore {
		sb.WriteString("(more results available \u2014 increase --limit to see them)\n")
	}
	return sb.String()
}

func (f *lineFormatter) FormatError(err error) string {
	return formatErrorPlain(err, f.useColor)
}
