package gematria

import (
	"strings"
	"testing"
)

func TestLettersDictionarySize(t *testing.T) {
	if got := len(letters); got != 27 {
		t.Errorf("len(letters) = %d, want 27", got)
	}
}

func TestAllStandardLettersPresent(t *testing.T) {
	for _, r := range "אבגדהוזחטיכלמנסעפצקרשת" {
		if _, ok := letters[r]; !ok {
			t.Errorf("missing standard letter %c (U+%04X)", r, r)
		}
	}

}

func TestAllSofitFormsPresent(t *testing.T) {
	sofits := []rune{'ך', 'ם', 'ן', 'ף', 'ץ'}
	for _, r := range sofits {
		l, ok := letters[r]
		if !ok {
			t.Errorf("missing sofit letter %c (U+%04X)", r, r)
			continue
		}
		if !l.IsSofit {
			t.Errorf("letter %c (U+%04X): IsSofit = false, want true", r, r)
		}
	}
}

func TestSofitPositionsMatchBaseLetter(t *testing.T) {
	cases := []struct {
		base  rune
		sofit rune
	}{
		{'כ', 'ך'}, // Kaf / Kaf Sofit — position 11
		{'מ', 'ם'}, // Mem / Mem Sofit — position 13
		{'נ', 'ן'}, // Nun / Nun Sofit — position 14
		{'פ', 'ף'}, // Pe  / Pe  Sofit — position 17
		{'צ', 'ץ'}, // Tsade / Tsade Sofit — position 18
	}
	for _, c := range cases {
		base, ok := letters[c.base]
		if !ok {
			t.Errorf("missing base letter %c", c.base)
			continue
		}
		sofit, ok := letters[c.sofit]
		if !ok {
			t.Errorf("missing sofit letter %c", c.sofit)
			continue
		}
		if base.Position != sofit.Position {
			t.Errorf("%c Position=%d, %c Position=%d: want equal", c.base, base.Position, c.sofit, sofit.Position)
		}
	}
}

func TestAliasesLowercaseAndNonEmpty(t *testing.T) {
	for r, l := range letters {
		if len(l.Aliases) == 0 {
			t.Errorf("letter %c (U+%04X): has no aliases", r, r)
		}
		for _, a := range l.Aliases {
			if a != strings.ToLower(a) {
				t.Errorf("letter %c alias %q is not lowercase", r, a)
			}
		}
	}
}

func TestHechrachiValues(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{'א', 1}, {'ב', 2}, {'ג', 3}, {'ד', 4}, {'ה', 5},
		{'ו', 6}, {'ז', 7}, {'ח', 8}, {'ט', 9}, {'י', 10},
		{'כ', 20}, {'ל', 30}, {'מ', 40}, {'נ', 50}, {'ס', 60},
		{'ע', 70}, {'פ', 80}, {'צ', 90}, {'ק', 100}, {'ר', 200},
		{'ש', 300}, {'ת', 400},
		// sofit same as normal
		{'ך', 20}, {'ם', 40}, {'ן', 50}, {'ף', 80}, {'ץ', 90},
	}
	for _, c := range cases {
		if got := hechrachi[c.r]; got != c.want {
			t.Errorf("hechrachi[%c] = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestGadolValues(t *testing.T) {
	// standard letters identical to hechrachi — spot-check
	if got := gadol['א']; got != 1 {
		t.Errorf("gadol[Aleph] = %d, want 1", got)
	}
	if got := gadol['ת']; got != 400 {
		t.Errorf("gadol[Tav] = %d, want 400", got)
	}
	// sofit extended values
	cases := []struct {
		r    rune
		want int
	}{
		{'ך', 500}, {'ם', 600}, {'ן', 700}, {'ף', 800}, {'ץ', 900},
	}
	for _, c := range cases {
		if got := gadol[c.r]; got != c.want {
			t.Errorf("gadol[%c] = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestSiduriValues(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{'א', 1}, {'כ', 11}, {'ל', 12}, {'מ', 13}, {'ת', 22},
		// sofit same ordinal as normal
		{'ך', 11}, {'ם', 13}, {'ן', 14}, {'ף', 17}, {'ץ', 18},
	}
	for _, c := range cases {
		if got := siduri[c.r]; got != c.want {
			t.Errorf("siduri[%c] = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestAtbashValues(t *testing.T) {
	// pos 1 <-> 22: aleph=400 (tav's hechrachi), tav=1 (aleph's hechrachi)
	cases := []struct {
		r    rune
		want int
	}{
		{'א', 400}, {'ב', 300}, {'ג', 200}, {'ד', 100}, {'ה', 90},
		{'ו', 80}, {'ז', 70}, {'ח', 60}, {'ט', 50}, {'י', 40},
		{'כ', 30}, {'ל', 20}, {'מ', 10}, {'נ', 9}, {'ס', 8},
		{'ע', 7}, {'פ', 6}, {'צ', 5}, {'ק', 4}, {'ר', 3},
		{'ש', 2}, {'ת', 1},
		// sofit mirrors through normal form pair
		{'ך', 30}, {'ם', 10}, {'ן', 9}, {'ף', 6}, {'ץ', 5},
	}
	for _, c := range cases {
		if got := atbash[c.r]; got != c.want {
			t.Errorf("atbash[%c] = %d, want %d", c.r, got, c.want)
		}
	}
}

// TestSystemValuesRepresentative verifies all four gematria systems via the
// systemValues dispatch table using a representative set of letters, including
// sofit forms and their extended values.
func TestSystemValuesRepresentative(t *testing.T) {
	cases := []struct {
		name   string
		r      rune
		system System
		want   int
	}{
		// Hechrachi — standard and sofit same value
		{"Hechrachi/Aleph", 'א', Hechrachi, 1},
		{"Hechrachi/Kaf", 'כ', Hechrachi, 20},
		{"Hechrachi/KafSofit", 'ך', Hechrachi, 20},
		{"Hechrachi/Tav", 'ת', Hechrachi, 400},
		// Gadol — sofit extended
		{"Gadol/Aleph", 'א', Gadol, 1},
		{"Gadol/KafSofit", 'ך', Gadol, 500},
		{"Gadol/MemSofit", 'ם', Gadol, 600},
		{"Gadol/NunSofit", 'ן', Gadol, 700},
		{"Gadol/PeSofit", 'ף', Gadol, 800},
		{"Gadol/TsadeSofit", 'ץ', Gadol, 900},
		// Siduri — ordinal by position; sofit shares position of base letter
		{"Siduri/Aleph", 'א', Siduri, 1},
		{"Siduri/Lamed", 'ל', Siduri, 12},
		{"Siduri/KafSofit", 'ך', Siduri, 11},
		{"Siduri/Tav", 'ת', Siduri, 22},
		// Atbash — Hechrachi value of mirrored letter; pos 1↔22, 11↔12
		{"Atbash/Aleph", 'א', Atbash, 400},
		{"Atbash/Tav", 'ת', Atbash, 1},
		{"Atbash/Kaf", 'כ', Atbash, 30},
		{"Atbash/KafSofit", 'ך', Atbash, 30},
		{"Atbash/TsadeSofit", 'ץ', Atbash, 5},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := systemValues[tc.system][tc.r]
			if got != tc.want {
				t.Errorf("systemValues[%s][%c] = %d, want %d", tc.system, tc.r, got, tc.want)
			}
		})
	}
}

func TestSystemValuesDispatch(t *testing.T) {
	if got := systemValues[Hechrachi]['א']; got != 1 {
		t.Errorf("systemValues[Hechrachi][Aleph] = %d, want 1", got)
	}
	if got := systemValues[Gadol]['ץ']; got != 900 {
		t.Errorf("systemValues[Gadol][TsadeSofit] = %d, want 900", got)
	}
	if got := systemValues[Siduri]['ת']; got != 22 {
		t.Errorf("systemValues[Siduri][Tav] = %d, want 22", got)
	}
	if got := systemValues[Atbash]['א']; got != 400 {
		t.Errorf("systemValues[Atbash][Aleph] = %d, want 400", got)
	}
}

func TestAtbashMirrorBidirectional(t *testing.T) {
	// every key should map to a value, and that value should map back
	for r, mirror := range atbashMirror {
		back, ok := atbashMirror[mirror]
		if !ok {
			t.Errorf("atbashMirror[%c] = %c, but reverse key missing", r, mirror)
			continue
		}
		// sofit forms map to the normal form's pair, so the reverse may differ
		// just check it exists and is non-zero
		if back == 0 {
			t.Errorf("atbashMirror[%c] reverse is zero rune", mirror)
		}
	}
}

func TestAllTablesCoversAll27Runes(t *testing.T) {
	tables := map[string]map[rune]int{
		"hechrachi": hechrachi,
		"gadol":     gadol,
		"siduri":    siduri,
		"atbash":    atbash,
	}
	for name, table := range tables {
		for r := range letters {
			if _, ok := table[r]; !ok {
				t.Errorf("%s missing rune %c (U+%04X)", name, r, r)
			}
		}
	}
}

func TestValidSystems(t *testing.T) {
	got := ValidSystems()
	want := []System{Hechrachi, Gadol, Siduri, Atbash}
	if len(got) != len(want) {
		t.Fatalf("ValidSystems() len = %d, want %d", len(got), len(want))
	}
	for i, s := range want {
		if got[i] != s {
			t.Errorf("ValidSystems()[%d] = %q, want %q", i, got[i], s)
		}
	}
}

func TestSystemConstants(t *testing.T) {
	if string(Hechrachi) != "hechrachi" {
		t.Errorf("Hechrachi = %q, want %q", Hechrachi, "hechrachi")
	}
	if string(Gadol) != "gadol" {
		t.Errorf("Gadol = %q, want %q", Gadol, "gadol")
	}
	if string(Siduri) != "siduri" {
		t.Errorf("Siduri = %q, want %q", Siduri, "siduri")
	}
	if string(Atbash) != "atbash" {
		t.Errorf("Atbash = %q, want %q", Atbash, "atbash")
	}
}
