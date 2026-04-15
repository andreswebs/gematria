package gematria

import "testing"

// --- LetterResult ---

// TestLetterResultEmbeddingPromotesLetterFields verifies that Letter fields
// are promoted directly onto LetterResult (no indirection needed).
func TestLetterResultEmbeddingPromotesLetterFields(t *testing.T) {
	lr := LetterResult{
		Letter: Letter{
			Char:     'א',
			Name:     "Aleph",
			Meaning:  "ox",
			Position: 1,
		},
		Value: 1,
	}

	if lr.Char != 'א' {
		t.Errorf("Char = %c, want א", lr.Char)
	}
	if lr.Name != "Aleph" {
		t.Errorf("Name = %q, want Aleph", lr.Name)
	}
	if lr.Meaning != "ox" {
		t.Errorf("Meaning = %q, want ox", lr.Meaning)
	}
	if lr.Position != 1 {
		t.Errorf("Position = %d, want 1", lr.Position)
	}
	if lr.Value != 1 {
		t.Errorf("Value = %d, want 1", lr.Value)
	}
}

// --- Result ---

// TestResultFields verifies the Result struct carries all required fields.
func TestResultFields(t *testing.T) {
	r := Result{
		Input:  "אבג",
		System: Hechrachi,
		Total:  6,
		Letters: []LetterResult{
			{Letter: letters['א'], Value: 1},
			{Letter: letters['ב'], Value: 2},
			{Letter: letters['ג'], Value: 3},
		},
	}

	if r.Input != "אבג" {
		t.Errorf("Input = %q, want אבג", r.Input)
	}
	if r.System != Hechrachi {
		t.Errorf("System = %q, want %q", r.System, Hechrachi)
	}
	if r.Total != 6 {
		t.Errorf("Total = %d, want 6", r.Total)
	}
	if len(r.Letters) != 3 {
		t.Fatalf("len(Letters) = %d, want 3", len(r.Letters))
	}
	// Verify promoted fields on the first LetterResult.
	if r.Letters[0].Name != "Aleph" {
		t.Errorf("Letters[0].Name = %q, want Aleph", r.Letters[0].Name)
	}
	if r.Letters[0].Value != 1 {
		t.Errorf("Letters[0].Value = %d, want 1", r.Letters[0].Value)
	}
}

// TestResultZeroValue verifies zero-value Result is usable without panics.
func TestResultZeroValue(t *testing.T) {
	var r Result
	if r.Total != 0 {
		t.Errorf("zero Total = %d, want 0", r.Total)
	}
	if len(r.Letters) != 0 {
		t.Errorf("zero Letters len = %d, want 0", len(r.Letters))
	}
	if r.Scheme != "" {
		t.Errorf("zero Scheme = %q, want empty", r.Scheme)
	}
}

// TestResultSchemeFieldEmptyForCompute verifies that Result.Scheme is empty
// when Compute (no transliteration) produces the result.
func TestResultSchemeFieldEmptyForCompute(t *testing.T) {
	r, err := Compute("שלום", Hechrachi)
	if err != nil {
		t.Fatalf("Compute unexpected error: %v", err)
	}
	if r.Scheme != "" {
		t.Errorf("Compute().Scheme = %q, want empty", r.Scheme)
	}
}

// TestResultSchemeFieldSetForTransliterated verifies that ComputeTransliterated
// populates Scheme on the returned Result.
func TestResultSchemeFieldSetForTransliterated(t *testing.T) {
	cases := []struct {
		input  string
		scheme Scheme
	}{
		{"shalom", SchemeAcademic},
		{"shalom", SchemeIsraeli},
		{"gadol", SchemeAcademic},
		{"gadol", SchemeIsraeli},
	}
	for _, tc := range cases {
		t.Run(tc.input+"_"+string(tc.scheme), func(t *testing.T) {
			r, err := ComputeTransliterated(tc.input, Hechrachi, tc.scheme)
			if err != nil {
				t.Fatalf("ComputeTransliterated(%q, %q) unexpected error: %v", tc.input, tc.scheme, err)
			}
			if r.Scheme != tc.scheme {
				t.Errorf("ComputeTransliterated(%q, %q).Scheme = %q, want %q",
					tc.input, tc.scheme, r.Scheme, tc.scheme)
			}
		})
	}
}
