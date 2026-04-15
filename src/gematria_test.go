package gematria

import (
	"errors"
	"testing"
)

// --- LookupLetter() ---

// TestLookupLetterAllEntries verifies all 27 dictionary entries through the
// public LookupLetter API, checking Name, Meaning, Position, IsSofit, and
// that Aliases is non-empty.
func TestLookupLetterAllEntries(t *testing.T) {
	cases := []struct {
		char    rune
		name    string
		meaning string
		pos     int
		isSofit bool
	}{
		// 22 standard letters
		{'א', "Aleph", "ox", 1, false},
		{'ב', "Bet", "house", 2, false},
		{'ג', "Gimel", "camel", 3, false},
		{'ד', "Dalet", "door", 4, false},
		{'ה', "He", "window", 5, false},
		{'ו', "Vav", "hook", 6, false},
		{'ז', "Zayin", "weapon", 7, false},
		{'ח', "Het", "enclosure", 8, false},
		{'ט', "Tet", "serpent", 9, false},
		{'י', "Yud", "hand", 10, false},
		{'כ', "Kaf", "palm", 11, false},
		{'ל', "Lamed", "goad", 12, false},
		{'מ', "Mem", "water", 13, false},
		{'נ', "Nun", "fish", 14, false},
		{'ס', "Samekh", "support", 15, false},
		{'ע', "Ayin", "eye", 16, false},
		{'פ', "Pe", "mouth", 17, false},
		{'צ', "Tsade", "fish hook", 18, false},
		{'ק', "Qoph", "eye of needle", 19, false},
		{'ר', "Resh", "head", 20, false},
		{'ש', "Shin", "tooth", 21, false},
		{'ת', "Tav", "cross", 22, false},
		// 5 sofit (final) forms — share Position with their base letter
		{'ך', "Kaf Sofit", "palm", 11, true},
		{'ם', "Mem Sofit", "water", 13, true},
		{'ן', "Nun Sofit", "fish", 14, true},
		{'ף', "Pe Sofit", "mouth", 17, true},
		{'ץ', "Tsade Sofit", "fish hook", 18, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := LookupLetter(tc.char)
			if err != nil {
				t.Fatalf("LookupLetter(%c) returned error: %v", tc.char, err)
			}
			if got.Name != tc.name {
				t.Errorf("Name = %q, want %q", got.Name, tc.name)
			}
			if got.Meaning != tc.meaning {
				t.Errorf("Meaning = %q, want %q", got.Meaning, tc.meaning)
			}
			if got.Position != tc.pos {
				t.Errorf("Position = %d, want %d", got.Position, tc.pos)
			}
			if got.IsSofit != tc.isSofit {
				t.Errorf("IsSofit = %v, want %v", got.IsSofit, tc.isSofit)
			}
			if len(got.Aliases) == 0 {
				t.Errorf("Aliases is empty, want at least one alias")
			}
		})
	}
}

func TestLookupLetterKnownRune(t *testing.T) {
	l, err := LookupLetter('א')
	if err != nil {
		t.Fatalf("LookupLetter('א') returned error: %v", err)
	}
	if l.Char != 'א' {
		t.Errorf("Char = %c, want א", l.Char)
	}
	if l.Name != "Aleph" {
		t.Errorf("Name = %q, want %q", l.Name, "Aleph")
	}
	if l.Position != 1 {
		t.Errorf("Position = %d, want 1", l.Position)
	}
}

func TestLookupLetterUnknownRune(t *testing.T) {
	_, err := LookupLetter('X')
	if err == nil {
		t.Fatal("LookupLetter('X') returned nil error, want *InvalidCharError")
	}
	var ice *InvalidCharError
	if !errors.As(err, &ice) {
		t.Errorf("err type = %T, want *InvalidCharError", err)
	}
	if ice.Char != 'X' {
		t.Errorf("InvalidCharError.Char = %c, want X", ice.Char)
	}
}

// --- LetterByName() ---

// TestLetterByNameAliasVariants verifies that multiple aliases per letter
// (including sofit forms) resolve to the correct rune, and that the lookup
// is case-insensitive.
func TestLetterByNameAliasVariants(t *testing.T) {
	cases := []struct {
		input    string
		wantChar rune
	}{
		// aleph — two canonical aliases, plus case variants
		{"aleph", 'א'}, {"alef", 'א'}, {"ALEPH", 'א'}, {"Aleph", 'א'},
		// bet — three aliases
		{"bet", 'ב'}, {"beth", 'ב'}, {"vet", 'ב'},
		// gimel — two aliases
		{"gimel", 'ג'}, {"gamel", 'ג'},
		// dalet — two aliases
		{"dalet", 'ד'}, {"daleth", 'ד'},
		// he — two aliases
		{"he", 'ה'}, {"heh", 'ה'},
		// vav — three aliases
		{"vav", 'ו'}, {"vau", 'ו'}, {"waw", 'ו'},
		// het — three aliases
		{"het", 'ח'}, {"chet", 'ח'}, {"khet", 'ח'},
		// kaf — multiple aliases
		{"kaf", 'כ'}, {"kaph", 'כ'}, {"chaf", 'כ'},
		// shin — two aliases
		{"shin", 'ש'}, {"sin", 'ש'},
		// tav — two aliases
		{"tav", 'ת'}, {"taw", 'ת'},
		// sofit aliases — two per sofit letter
		{"kaf sofit", 'ך'}, {"final kaf", 'ך'},
		{"mem sofit", 'ם'}, {"final mem", 'ם'},
		{"nun sofit", 'ן'}, {"final nun", 'ן'},
		{"pe sofit", 'ף'}, {"final pe", 'ף'},
		{"tsade sofit", 'ץ'}, {"final tsade", 'ץ'},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got, err := LetterByName(tc.input)
			if err != nil {
				t.Fatalf("LetterByName(%q) returned error: %v", tc.input, err)
			}
			if got.Char != tc.wantChar {
				t.Errorf("Char = %c (U+%04X), want %c (U+%04X)",
					got.Char, got.Char, tc.wantChar, tc.wantChar)
			}
		})
	}
}

func TestLetterByNameExactAlias(t *testing.T) {
	l, err := LetterByName("aleph")
	if err != nil {
		t.Fatalf("LetterByName(%q) returned error: %v", "aleph", err)
	}
	if l.Char != 'א' {
		t.Errorf("Char = %c, want א", l.Char)
	}
}

func TestLetterByNameCaseInsensitive(t *testing.T) {
	for _, input := range []string{"ALEPH", "Aleph", "aLePh"} {
		l, err := LetterByName(input)
		if err != nil {
			t.Fatalf("LetterByName(%q) returned error: %v", input, err)
		}
		if l.Char != 'א' {
			t.Errorf("LetterByName(%q).Char = %c, want א", input, l.Char)
		}
	}
}

func TestLetterByNameUnknownWithSuggestion(t *testing.T) {
	// "shen" is edit distance 1 from "shin"
	_, err := LetterByName("shen")
	if err == nil {
		t.Fatal("LetterByName(\"shen\") returned nil error, want *UnknownNameError")
	}
	var une *UnknownNameError
	if !errors.As(err, &une) {
		t.Errorf("err type = %T, want *UnknownNameError", err)
	}
	if une.Name != "shen" {
		t.Errorf("UnknownNameError.Name = %q, want %q", une.Name, "shen")
	}
	found := false
	for _, s := range une.Suggestions {
		if s == "shin" {
			found = true
		}
	}
	if !found {
		t.Errorf("Suggestions = %v, want to contain %q", une.Suggestions, "shin")
	}
}

func TestLetterByNameUnknownNoSuggestions(t *testing.T) {
	// "xyzzy" is well beyond edit distance 2 from any alias
	_, err := LetterByName("xyzzy")
	if err == nil {
		t.Fatal("LetterByName(\"xyzzy\") returned nil error, want *UnknownNameError")
	}
	var une *UnknownNameError
	if !errors.As(err, &une) {
		t.Errorf("err type = %T, want *UnknownNameError", err)
	}
	if len(une.Suggestions) != 0 {
		t.Errorf("Suggestions = %v, want empty for distant input", une.Suggestions)
	}
}

// --- AtbashSubstitute() ---

// TestAtbashSubstitutePairs verifies specific mirror pairs including sofit
// forms using a table-driven approach. Sofit forms substitute to the normal
// form's pair rune (e.g., kaf sofit → lamed, same as kaf → lamed).
func TestAtbashSubstitutePairs(t *testing.T) {
	cases := []struct {
		name   string
		r      rune
		mirror rune
	}{
		// standard bidirectional pairs
		{"aleph-tav", 'א', 'ת'},
		{"tav-aleph", 'ת', 'א'},
		{"bet-shin", 'ב', 'ש'},
		{"shin-bet", 'ש', 'ב'},
		{"kaf-lamed", 'כ', 'ל'},
		{"lamed-kaf", 'ל', 'כ'},
		// sofit forms → normal form's pair rune
		{"kaf-sofit-lamed", 'ך', 'ל'},
		{"mem-sofit-yud", 'ם', 'י'},
		{"nun-sofit-tet", 'ן', 'ט'},
		{"pe-sofit-vav", 'ף', 'ו'},
		{"tsade-sofit-he", 'ץ', 'ה'},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := AtbashSubstitute(tc.r); got != tc.mirror {
				t.Errorf("AtbashSubstitute(%c) = %c, want %c", tc.r, got, tc.mirror)
			}
		})
	}
}

func TestAtbashSubstituteKnownRune(t *testing.T) {
	// Aleph (position 1) <-> Tav (position 22)
	if got := AtbashSubstitute('א'); got != 'ת' {
		t.Errorf("AtbashSubstitute('א') = %c, want ת", got)
	}
	if got := AtbashSubstitute('ת'); got != 'א' {
		t.Errorf("AtbashSubstitute('ת') = %c, want א", got)
	}
}

func TestAtbashSubstituteRoundtrip(t *testing.T) {
	for r := range atbashMirror {
		once := AtbashSubstitute(r)
		back := AtbashSubstitute(once)
		if back == 0 {
			t.Errorf("AtbashSubstitute(AtbashSubstitute(%c)) = zero rune", r)
		}
	}
}

func TestAtbashSubstituteUnknownRunePassthrough(t *testing.T) {
	if got := AtbashSubstitute('X'); got != 'X' {
		t.Errorf("AtbashSubstitute('X') = %c, want X (passthrough)", got)
	}
}

// --- Compute() ---

func TestComputeHechrachi_Aleph(t *testing.T) {
	res, err := Compute("א", Hechrachi)
	if err != nil {
		t.Fatalf("Compute(%q, Hechrachi) returned error: %v", "א", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
	if len(res.Letters) != 1 {
		t.Fatalf("len(Letters) = %d, want 1", len(res.Letters))
	}
	if res.Letters[0].Value != 1 {
		t.Errorf("Letters[0].Value = %d, want 1", res.Letters[0].Value)
	}
}

func TestComputeHechrachi_Shalom(t *testing.T) {
	// שלום = shin(300) + lamed(30) + vav(6) + mem-sofit(40) = 376
	res, err := Compute("שלום", Hechrachi)
	if err != nil {
		t.Fatalf("Compute(\"שלום\", Hechrachi) error: %v", err)
	}
	if res.Total != 376 {
		t.Errorf("Total = %d, want 376", res.Total)
	}
	if res.Input != "שלום" {
		t.Errorf("Input = %q, want %q", res.Input, "שלום")
	}
	if res.System != Hechrachi {
		t.Errorf("System = %q, want %q", res.System, Hechrachi)
	}
	if len(res.Letters) != 4 {
		t.Fatalf("len(Letters) = %d, want 4", len(res.Letters))
	}
}

func TestComputeAllSystems_Aleph(t *testing.T) {
	cases := []struct {
		system System
		want   int
	}{
		{Hechrachi, 1},
		{Gadol, 1},
		{Siduri, 1},
		{Atbash, 400}, // aleph mirrors tav (hechrachi 400)
	}
	for _, tc := range cases {
		res, err := Compute("א", tc.system)
		if err != nil {
			t.Errorf("Compute(aleph, %s) error: %v", tc.system, err)
			continue
		}
		if res.Total != tc.want {
			t.Errorf("Compute(aleph, %s).Total = %d, want %d", tc.system, res.Total, tc.want)
		}
	}
}

func TestComputeInvalidSystem(t *testing.T) {
	_, err := Compute("א", System("bogus"))
	if err == nil {
		t.Fatal("expected error for invalid system, got nil")
	}
	var ise *InvalidSystemError
	if !errors.As(err, &ise) {
		t.Errorf("err type = %T, want *InvalidSystemError", err)
	}
	if ise.Name != "bogus" {
		t.Errorf("InvalidSystemError.Name = %q, want %q", ise.Name, "bogus")
	}
}

func TestComputeAtbash_OriginalCharPreserved(t *testing.T) {
	// Atbash: aleph (position 1) mirrors tav (position 22, value 400)
	// LetterResult.Char should still be aleph (original input)
	res, err := Compute("א", Atbash)
	if err != nil {
		t.Fatalf("Compute(aleph, Atbash) error: %v", err)
	}
	if res.Letters[0].Char != 'א' {
		t.Errorf("LetterResult.Char = %c, want א (original preserved)", res.Letters[0].Char)
	}
	if res.Letters[0].Value != 400 {
		t.Errorf("LetterResult.Value = %d, want 400 (atbash value)", res.Letters[0].Value)
	}
}

// --- Parse() ---

func TestParsePureHebrew(t *testing.T) {
	ltrs, err := Parse("אבג")
	if err != nil {
		t.Fatalf("Parse(\"אבג\") error: %v", err)
	}
	if len(ltrs) != 3 {
		t.Fatalf("len = %d, want 3", len(ltrs))
	}
	names := []string{"Aleph", "Bet", "Gimel"}
	for i, name := range names {
		if ltrs[i].Name != name {
			t.Errorf("ltrs[%d].Name = %q, want %q", i, ltrs[i].Name, name)
		}
	}
}

func TestParsePureHebrewWithSpaces(t *testing.T) {
	ltrs, err := Parse("א ב")
	if err != nil {
		t.Fatalf("Parse(\"א ב\") error: %v", err)
	}
	if len(ltrs) != 2 {
		t.Fatalf("len = %d, want 2", len(ltrs))
	}
}

func TestParsePureLatin(t *testing.T) {
	ltrs, err := Parse("aleph bet")
	if err != nil {
		t.Fatalf("Parse(\"aleph bet\") error: %v", err)
	}
	if len(ltrs) != 2 {
		t.Fatalf("len = %d, want 2", len(ltrs))
	}
	if ltrs[0].Char != 'א' {
		t.Errorf("ltrs[0].Char = %c, want א", ltrs[0].Char)
	}
	if ltrs[1].Char != 'ב' {
		t.Errorf("ltrs[1].Char = %c, want ב", ltrs[1].Char)
	}
}

func TestParseMixed(t *testing.T) {
	// "aleph ב" — first token is Latin, second is Hebrew rune
	ltrs, err := Parse("aleph ב")
	if err != nil {
		t.Fatalf("Parse(\"aleph ב\") error: %v", err)
	}
	if len(ltrs) != 2 {
		t.Fatalf("len = %d, want 2", len(ltrs))
	}
	if ltrs[0].Char != 'א' {
		t.Errorf("ltrs[0].Char = %c, want א", ltrs[0].Char)
	}
	if ltrs[1].Char != 'ב' {
		t.Errorf("ltrs[1].Char = %c, want ב", ltrs[1].Char)
	}
}

func TestComputeLatinInput(t *testing.T) {
	// "aleph" in Hechrachi → 1
	res, err := Compute("aleph", Hechrachi)
	if err != nil {
		t.Fatalf("Compute(\"aleph\", Hechrachi) error: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1", res.Total)
	}
}

func TestComputeGadol_SofitExtended(t *testing.T) {
	// כ (kaf normal) = 20 in Hechrachi, same in Gadol
	// ך (kaf sofit) = 20 in Hechrachi, 500 in Gadol
	res, err := Compute("ך", Gadol)
	if err != nil {
		t.Fatalf("Compute(kaf sofit, Gadol) error: %v", err)
	}
	if res.Total != 500 {
		t.Errorf("Total = %d, want 500 (kaf sofit in Gadol)", res.Total)
	}
}

func TestComputeSiduri_Ordinal(t *testing.T) {
	// Aleph = position 1 in Siduri
	res, err := Compute("א", Siduri)
	if err != nil {
		t.Fatalf("Compute(aleph, Siduri) error: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1 (Siduri ordinal)", res.Total)
	}
	// Tav = position 22 in Siduri
	res2, err := Compute("ת", Siduri)
	if err != nil {
		t.Fatalf("Compute(tav, Siduri) error: %v", err)
	}
	if res2.Total != 22 {
		t.Errorf("Total = %d, want 22 (Tav Siduri ordinal)", res2.Total)
	}
}

func TestComputeMixedInput(t *testing.T) {
	// "aleph ב" mixed → aleph(1) + bet(2) = 3 in Hechrachi
	res, err := Compute("aleph ב", Hechrachi)
	if err != nil {
		t.Fatalf("Compute(\"aleph ב\", Hechrachi) error: %v", err)
	}
	if res.Total != 3 {
		t.Errorf("Total = %d, want 3", res.Total)
	}
	if len(res.Letters) != 2 {
		t.Errorf("len(Letters) = %d, want 2", len(res.Letters))
	}
}

func TestParseEmpty(t *testing.T) {
	ltrs, err := Parse("")
	if err != nil {
		t.Fatalf("Parse(\"\") error: %v", err)
	}
	if len(ltrs) != 0 {
		t.Errorf("len = %d, want 0", len(ltrs))
	}
}

func TestParseLatin_UnknownName(t *testing.T) {
	_, err := Parse("aleph xyzzy")
	if err == nil {
		t.Fatal("expected error for unknown name, got nil")
	}
	var une *UnknownNameError
	if !errors.As(err, &une) {
		t.Errorf("err type = %T, want *UnknownNameError", err)
	}
	if une.Name != "xyzzy" {
		t.Errorf("UnknownNameError.Name = %q, want %q", une.Name, "xyzzy")
	}
	if une.Position != 1 {
		t.Errorf("UnknownNameError.Position = %d, want 1 (second token)", une.Position)
	}
}

// TestParseAllHebrewLetters verifies Parse() with all 27 Hebrew runes individually.
func TestParseAllHebrewLetters(t *testing.T) {
	cases := []struct {
		r    rune
		name string
	}{
		{'א', "Aleph"}, {'ב', "Bet"}, {'ג', "Gimel"}, {'ד', "Dalet"}, {'ה', "He"},
		{'ו', "Vav"}, {'ז', "Zayin"}, {'ח', "Het"}, {'ט', "Tet"}, {'י', "Yud"},
		{'כ', "Kaf"}, {'ל', "Lamed"}, {'מ', "Mem"}, {'נ', "Nun"}, {'ס', "Samekh"},
		{'ע', "Ayin"}, {'פ', "Pe"}, {'צ', "Tsade"}, {'ק', "Qoph"}, {'ר', "Resh"},
		{'ש', "Shin"}, {'ת', "Tav"},
		{'ך', "Kaf Sofit"}, {'ם', "Mem Sofit"}, {'ן', "Nun Sofit"},
		{'ף', "Pe Sofit"}, {'ץ', "Tsade Sofit"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ltrs, err := Parse(string(tc.r))
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", string(tc.r), err)
			}
			if len(ltrs) != 1 {
				t.Fatalf("Parse(%q): got %d letters, want 1", string(tc.r), len(ltrs))
			}
			if ltrs[0].Char != tc.r {
				t.Errorf("Parse(%q): Char = %c, want %c", string(tc.r), ltrs[0].Char, tc.r)
			}
			if ltrs[0].Name != tc.name {
				t.Errorf("Parse(%q): Name = %q, want %q", string(tc.r), ltrs[0].Name, tc.name)
			}
		})
	}
}

// TestParseLatin_MultiToken verifies that space-separated Latin names resolve
// to the correct letter sequence.
func TestParseLatin_MultiToken(t *testing.T) {
	// "aleph mem tav" → א + מ + ת
	ltrs, err := Parse("aleph mem tav")
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", "aleph mem tav", err)
	}
	if len(ltrs) != 3 {
		t.Fatalf("Parse(%q): got %d letters, want 3", "aleph mem tav", len(ltrs))
	}
	want := []rune{'א', 'מ', 'ת'}
	for i, r := range want {
		if ltrs[i].Char != r {
			t.Errorf("ltrs[%d].Char = %c, want %c", i, ltrs[i].Char, r)
		}
	}
}

// TestParseLatin_CaseInsensitive verifies that Latin input is case-insensitive via Parse().
func TestParseLatin_CaseInsensitive(t *testing.T) {
	for _, input := range []string{"ALEPH", "Aleph", "aLePh"} {
		input := input
		t.Run(input, func(t *testing.T) {
			ltrs, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", input, err)
			}
			if len(ltrs) != 1 || ltrs[0].Char != 'א' {
				t.Errorf("Parse(%q): got %v, want aleph (א)", input, ltrs)
			}
		})
	}
}

// TestComputeHechrachi_Emet verifies "אמת" = 441 (aleph=1, mem=40, tav=400).
func TestComputeHechrachi_Emet(t *testing.T) {
	res, err := Compute("אמת", Hechrachi)
	if err != nil {
		t.Fatalf("Compute(%q, Hechrachi) error: %v", "אמת", err)
	}
	if res.Total != 441 {
		t.Errorf("Total = %d, want 441", res.Total)
	}
	if len(res.Letters) != 3 {
		t.Fatalf("len(Letters) = %d, want 3", len(res.Letters))
	}
}

// TestComputeGadol_AllSofitValues verifies the extended sofit values (500–900) in Gadol.
func TestComputeGadol_AllSofitValues(t *testing.T) {
	cases := []struct {
		name  string
		r     rune
		value int
	}{
		{"kaf-sofit", 'ך', 500},
		{"mem-sofit", 'ם', 600},
		{"nun-sofit", 'ן', 700},
		{"pe-sofit", 'ף', 800},
		{"tsade-sofit", 'ץ', 900},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := Compute(string(tc.r), Gadol)
			if err != nil {
				t.Fatalf("Compute(%c, Gadol) error: %v", tc.r, err)
			}
			if res.Total != tc.value {
				t.Errorf("Compute(%c, Gadol).Total = %d, want %d", tc.r, res.Total, tc.value)
			}
		})
	}
}

// TestComputeAtbash_Tav verifies tav (position 22) mirrors aleph in atbash (value 1).
func TestComputeAtbash_Tav(t *testing.T) {
	res, err := Compute("ת", Atbash)
	if err != nil {
		t.Fatalf("Compute(tav, Atbash) error: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("Total = %d, want 1 (tav mirrors aleph in atbash)", res.Total)
	}
	if res.Letters[0].Char != 'ת' {
		t.Errorf("LetterResult.Char = %c, want ת (original preserved)", res.Letters[0].Char)
	}
}

// TestComputeResultTotalInvariant verifies Result.Total == sum of all LetterResult.Value
// entries across multiple inputs and systems.
func TestComputeResultTotalInvariant(t *testing.T) {
	cases := []struct {
		input  string
		system System
	}{
		{"אמת", Hechrachi},
		{"שלום", Hechrachi},
		{"אמת", Siduri},
		{"ךםןףץ", Gadol},
		{"אבגד", Atbash},
		{"aleph mem tav", Hechrachi},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input+"/"+string(tc.system), func(t *testing.T) {
			res, err := Compute(tc.input, tc.system)
			if err != nil {
				t.Fatalf("Compute(%q, %s) error: %v", tc.input, tc.system, err)
			}
			sum := 0
			for _, lr := range res.Letters {
				sum += lr.Value
			}
			if res.Total != sum {
				t.Errorf("Compute(%q, %s): Total = %d, want sum of letters = %d",
					tc.input, tc.system, res.Total, sum)
			}
		})
	}
}

// --- ComputeFromLetters and ComputeTransliterated ---

// TestComputeFromLetters_DirectInvocation verifies the primitive can be
// invoked with a hand-built letter slice and produces the same result as
// Compute on equivalent input.
func TestComputeFromLetters_DirectInvocation(t *testing.T) {
	ltrs := []Letter{letters['א'], letters['מ'], letters['ת']}
	r, err := ComputeFromLetters("manual", ltrs, Hechrachi)
	if err != nil {
		t.Fatalf("ComputeFromLetters error: %v", err)
	}
	if r.Total != 441 {
		t.Errorf("ComputeFromLetters total = %d, want 441", r.Total)
	}
	if r.Input != "manual" {
		t.Errorf("ComputeFromLetters input = %q, want %q", r.Input, "manual")
	}
	if r.Scheme != "" {
		t.Errorf("ComputeFromLetters scheme = %q, want empty", r.Scheme)
	}
}

// TestComputeFromLetters_AgreesWithCompute verifies that Compute and
// ComputeFromLetters produce identical Results when given the same letters.
func TestComputeFromLetters_AgreesWithCompute(t *testing.T) {
	cases := []struct {
		input  string
		system System
	}{
		{"א", Hechrachi},
		{"שלום", Hechrachi},
		{"אמת", Gadol},
		{"דעת", Siduri},
		{"שלום", Atbash},
	}
	for _, tc := range cases {
		t.Run(tc.input+"/"+string(tc.system), func(t *testing.T) {
			rCompute, err := Compute(tc.input, tc.system)
			if err != nil {
				t.Fatalf("Compute error: %v", err)
			}
			ltrs, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			rFromLetters, err := ComputeFromLetters(tc.input, ltrs, tc.system)
			if err != nil {
				t.Fatalf("ComputeFromLetters error: %v", err)
			}
			if rCompute.Total != rFromLetters.Total {
				t.Errorf("Total mismatch: Compute=%d, ComputeFromLetters=%d",
					rCompute.Total, rFromLetters.Total)
			}
			if len(rCompute.Letters) != len(rFromLetters.Letters) {
				t.Errorf("Letters length mismatch: Compute=%d, ComputeFromLetters=%d",
					len(rCompute.Letters), len(rFromLetters.Letters))
			}
		})
	}
}

// TestComputeFromLetters_InvalidSystem verifies that the primitive validates
// the system parameter and returns *InvalidSystemError.
func TestComputeFromLetters_InvalidSystem(t *testing.T) {
	ltrs := []Letter{letters['א']}
	_, err := ComputeFromLetters("a", ltrs, System("bogus"))
	if err == nil {
		t.Fatal("ComputeFromLetters with bogus system: want error, got nil")
	}
	var ise *InvalidSystemError
	if !errors.As(err, &ise) {
		t.Errorf("ComputeFromLetters error type = %T, want *InvalidSystemError", err)
	}
}

// TestComputeTransliterated_AcademicShalom verifies the canonical academic
// computation: shalom → שלם (with mem-sofit) = 300+30+40 = 370.
func TestComputeTransliterated_AcademicShalom(t *testing.T) {
	r, err := ComputeTransliterated("shalom", Hechrachi, SchemeAcademic)
	if err != nil {
		t.Fatalf("ComputeTransliterated error: %v", err)
	}
	if r.Total != 370 {
		t.Errorf("ComputeTransliterated(shalom, hechrachi, academic).Total = %d, want 370", r.Total)
	}
	if r.Scheme != SchemeAcademic {
		t.Errorf("Scheme = %q, want %q", r.Scheme, SchemeAcademic)
	}
	if r.Input != "shalom" {
		t.Errorf("Input = %q, want %q", r.Input, "shalom")
	}
}

// TestComputeTransliterated_IsraeliShalom verifies the canonical israeli
// computation: shalom → שלום (with vav as o-mater, mem-sofit) = 300+30+6+40 = 376.
func TestComputeTransliterated_IsraeliShalom(t *testing.T) {
	r, err := ComputeTransliterated("shalom", Hechrachi, SchemeIsraeli)
	if err != nil {
		t.Fatalf("ComputeTransliterated error: %v", err)
	}
	if r.Total != 376 {
		t.Errorf("ComputeTransliterated(shalom, hechrachi, israeli).Total = %d, want 376", r.Total)
	}
	if r.Scheme != SchemeIsraeli {
		t.Errorf("Scheme = %q, want %q", r.Scheme, SchemeIsraeli)
	}
}

// TestComputeTransliterated_SchemesProduceDifferentValues verifies that
// the same input gives different gematria values under the two schemes.
func TestComputeTransliterated_SchemesProduceDifferentValues(t *testing.T) {
	cases := []struct {
		input        string
		wantAcademic int
		wantIsraeli  int
	}{
		{"shalom", 370, 376},
		{"gadol", 37, 43},
		{"emet", 440, 441},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			rA, err := ComputeTransliterated(tc.input, Hechrachi, SchemeAcademic)
			if err != nil {
				t.Fatalf("academic error: %v", err)
			}
			if rA.Total != tc.wantAcademic {
				t.Errorf("academic total = %d, want %d", rA.Total, tc.wantAcademic)
			}
			rI, err := ComputeTransliterated(tc.input, Hechrachi, SchemeIsraeli)
			if err != nil {
				t.Fatalf("israeli error: %v", err)
			}
			if rI.Total != tc.wantIsraeli {
				t.Errorf("israeli total = %d, want %d", rI.Total, tc.wantIsraeli)
			}
		})
	}
}

// TestComputeTransliterated_AtbashSystem verifies that --mispar atbash
// composes correctly with transliteration: shalom-israeli (שלום) under atbash
// = ש(2) + ל(20) + ו(80) + ם(10) = 112.
func TestComputeTransliterated_AtbashSystem(t *testing.T) {
	r, err := ComputeTransliterated("shalom", Atbash, SchemeIsraeli)
	if err != nil {
		t.Fatalf("ComputeTransliterated error: %v", err)
	}
	if r.Total != 112 {
		t.Errorf("ComputeTransliterated(shalom, atbash, israeli).Total = %d, want 112", r.Total)
	}
	if r.Scheme != SchemeIsraeli {
		t.Errorf("Scheme = %q, want israeli", r.Scheme)
	}
	if r.System != Atbash {
		t.Errorf("System = %q, want atbash", r.System)
	}
}

// TestComputeTransliterated_InvalidSchemePropagates verifies that an invalid
// scheme passed to ComputeTransliterated returns *InvalidSchemeError.
func TestComputeTransliterated_InvalidSchemePropagates(t *testing.T) {
	_, err := ComputeTransliterated("shalom", Hechrachi, Scheme("bogus"))
	if err == nil {
		t.Fatal("want *InvalidSchemeError, got nil")
	}
	var iscse *InvalidSchemeError
	if !errors.As(err, &iscse) {
		t.Errorf("error type = %T, want *InvalidSchemeError", err)
	}
}

// TestComputeTransliterated_UnknownWordPropagates verifies that unmappable
// input returns *UnknownWordError.
func TestComputeTransliterated_UnknownWordPropagates(t *testing.T) {
	_, err := ComputeTransliterated("h3llo", Hechrachi, SchemeAcademic)
	if err == nil {
		t.Fatal("want *UnknownWordError, got nil")
	}
	var uwe *UnknownWordError
	if !errors.As(err, &uwe) {
		t.Errorf("error type = %T, want *UnknownWordError", err)
	}
	if uwe.Scheme != SchemeAcademic {
		t.Errorf("uwe.Scheme = %q, want academic", uwe.Scheme)
	}
}
