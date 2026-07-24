package cli

import (
	"encoding/json"
	"errors"

	gematria "github.com/andreswebs/gematria"
)

// jsonFormatter renders results and errors as JSON objects.
// No ANSI codes or RTL marks are ever included in JSON output.
type jsonFormatter struct {
	findValue  int
	findSystem gematria.System
}

// jsonLetter is the per-letter JSON structure for FormatResult.
type jsonLetter struct {
	Char     string `json:"char"`
	Name     string `json:"name"`
	Value    int    `json:"value"`
	Meaning  string `json:"meaning"`
	Position int    `json:"position"`
}

// jsonResult is the top-level JSON structure for FormatResult.
// Scheme is omitted when transliteration was not used.
type jsonResult struct {
	Input   string       `json:"input"`
	System  string       `json:"system"`
	Scheme  string       `json:"scheme,omitempty"`
	Total   int          `json:"total"`
	Letters []jsonLetter `json:"letters"`
}

// jsonError is the JSON structure for FormatError.
// Line is omitted when zero (non-batch context).
// Scheme is omitted when the error is not transliteration-related.
type jsonError struct {
	Error        string   `json:"error"`
	Line         int      `json:"line,omitempty"`
	InvalidInput string   `json:"invalid_input"`
	Scheme       string   `json:"scheme,omitempty"`
	Position     int      `json:"position"`
	Suggestions  []string `json:"suggestions"`
}

func (f *jsonFormatter) FormatResult(r gematria.Result) string {
	letters := make([]jsonLetter, len(r.Letters))
	for i, lr := range r.Letters {
		letters[i] = jsonLetter{
			Char:     string(lr.Char),
			Name:     lr.Name,
			Value:    lr.Value,
			Meaning:  lr.Meaning,
			Position: lr.Position,
		}
	}
	out := jsonResult{
		Input:   r.Input,
		System:  string(r.System),
		Scheme:  string(r.Scheme),
		Total:   r.Total,
		Letters: letters,
	}
	b, _ := json.Marshal(out)
	return string(b) + "\n"
}

// jsonLookupWord is the per-word JSON structure for FormatLookup.
type jsonLookupWord struct {
	Word            string `json:"word"`
	Transliteration string `json:"transliteration,omitempty"`
	Meaning         string `json:"meaning,omitempty"`
}

// jsonLookup is the top-level JSON structure for FormatLookup.
type jsonLookup struct {
	Value   int              `json:"value"`
	System  string           `json:"system"`
	Results []jsonLookupWord `json:"results"`
	HasMore bool             `json:"hasMore"`
}

func (f *jsonFormatter) FormatLookup(words []gematria.Word, hasMore bool) string {
	results := make([]jsonLookupWord, 0, len(words))
	for _, w := range words {
		results = append(results, jsonLookupWord{
			Word:            w.Hebrew,
			Transliteration: w.Transliteration,
			Meaning:         w.Meaning,
		})
	}
	out := jsonLookup{
		Value:   f.findValue,
		System:  string(f.findSystem),
		Results: results,
		HasMore: hasMore,
	}
	b, _ := json.Marshal(out)
	return string(b) + "\n"
}

func (f *jsonFormatter) FormatError(err error) string {
	// Unwrap batch line context so the inner error drives field population.
	innerErr := err
	var ble *batchLineError
	if errors.As(err, &ble) {
		innerErr = ble.cause
	}

	je := jsonError{
		Error:       innerErr.Error(),
		Suggestions: []string{},
	}
	if ble != nil {
		je.Line = ble.lineNum
	}

	var ice *gematria.InvalidCharError
	var une *gematria.UnknownNameError
	var ise *gematria.InvalidSystemError
	var uwe *gematria.UnknownWordError
	var iscse *gematria.InvalidSchemeError

	switch {
	case errors.As(err, &ice):
		je.InvalidInput = string(ice.Char)
		je.Position = ice.Position
	case errors.As(err, &une):
		je.InvalidInput = une.Name
		je.Position = une.Position
		if une.Suggestions != nil {
			je.Suggestions = une.Suggestions
		}
	case errors.As(err, &uwe):
		je.InvalidInput = uwe.Input
		je.Position = uwe.Position
		je.Scheme = string(uwe.Scheme)
		if uwe.Suggestions != nil {
			je.Suggestions = uwe.Suggestions
		}
	case errors.As(err, &iscse):
		je.InvalidInput = iscse.Name
	case errors.As(err, &ise):
		je.InvalidInput = ise.Name
	}

	b, _ := json.Marshal(je)
	return string(b) + "\n"
}
