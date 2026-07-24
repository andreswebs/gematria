package gematria

// System is the gematria numbering system to use for computation.
// Constants match CLI flag values exactly.
type System string

const (
	Hechrachi System = "hechrachi"
	Gadol     System = "gadol"
	Siduri    System = "siduri"
	Atbash    System = "atbash"
)

// Letter holds the dictionary entry for a single Hebrew letter.
type Letter struct {
	Char     rune
	Name     string
	Meaning  string
	Position int
	Aliases  []string
	IsSofit  bool
}

// letters is the complete 27-entry Hebrew letter dictionary keyed by rune.
var letters = map[rune]Letter{
	// 22 standard letters
	'א': {Char: 'א', Name: "Aleph", Meaning: "ox", Position: 1, Aliases: []string{"aleph", "alef"}, IsSofit: false},
	'ב': {Char: 'ב', Name: "Bet", Meaning: "house", Position: 2, Aliases: []string{"bet", "beth", "vet"}, IsSofit: false},
	'ג': {Char: 'ג', Name: "Gimel", Meaning: "camel", Position: 3, Aliases: []string{"gimel", "gamel"}, IsSofit: false},
	'ד': {Char: 'ד', Name: "Dalet", Meaning: "door", Position: 4, Aliases: []string{"dalet", "daleth"}, IsSofit: false},
	'ה': {Char: 'ה', Name: "He", Meaning: "window", Position: 5, Aliases: []string{"he", "heh"}, IsSofit: false},
	'ו': {Char: 'ו', Name: "Vav", Meaning: "hook", Position: 6, Aliases: []string{"vav", "vau", "waw"}, IsSofit: false},
	'ז': {Char: 'ז', Name: "Zayin", Meaning: "weapon", Position: 7, Aliases: []string{"zayin", "zayn"}, IsSofit: false},
	'ח': {Char: 'ח', Name: "Het", Meaning: "enclosure", Position: 8, Aliases: []string{"het", "chet", "khet"}, IsSofit: false},
	'ט': {Char: 'ט', Name: "Tet", Meaning: "serpent", Position: 9, Aliases: []string{"tet", "teth"}, IsSofit: false},
	'י': {Char: 'י', Name: "Yud", Meaning: "hand", Position: 10, Aliases: []string{"yud", "yod", "yodh"}, IsSofit: false},
	'כ': {Char: 'כ', Name: "Kaf", Meaning: "palm", Position: 11, Aliases: []string{"kaf", "kaph", "chaf", "caf"}, IsSofit: false},
	'ל': {Char: 'ל', Name: "Lamed", Meaning: "goad", Position: 12, Aliases: []string{"lamed", "lam"}, IsSofit: false},
	'מ': {Char: 'מ', Name: "Mem", Meaning: "water", Position: 13, Aliases: []string{"mem"}, IsSofit: false},
	'נ': {Char: 'נ', Name: "Nun", Meaning: "fish", Position: 14, Aliases: []string{"nun"}, IsSofit: false},
	'ס': {Char: 'ס', Name: "Samekh", Meaning: "support", Position: 15, Aliases: []string{"samekh", "samech"}, IsSofit: false},
	'ע': {Char: 'ע', Name: "Ayin", Meaning: "eye", Position: 16, Aliases: []string{"ayin", "ain"}, IsSofit: false},
	'פ': {Char: 'פ', Name: "Pe", Meaning: "mouth", Position: 17, Aliases: []string{"pe", "peh", "fe", "feh"}, IsSofit: false},
	'צ': {Char: 'צ', Name: "Tsade", Meaning: "fish hook", Position: 18, Aliases: []string{"tsade", "tsadi", "tzade", "tzadi"}, IsSofit: false},
	'ק': {Char: 'ק', Name: "Qoph", Meaning: "eye of needle", Position: 19, Aliases: []string{"qoph", "koph", "qof", "kof", "kuf"}, IsSofit: false},
	'ר': {Char: 'ר', Name: "Resh", Meaning: "head", Position: 20, Aliases: []string{"resh"}, IsSofit: false},
	'ש': {Char: 'ש', Name: "Shin", Meaning: "tooth", Position: 21, Aliases: []string{"shin", "sin"}, IsSofit: false},
	'ת': {Char: 'ת', Name: "Tav", Meaning: "cross", Position: 22, Aliases: []string{"tav", "taw"}, IsSofit: false},

	// 5 sofit (final) forms — share Position with their base letter
	'ך': {Char: 'ך', Name: "Kaf Sofit", Meaning: "palm", Position: 11, Aliases: []string{"kaf sofit", "chaf sofit", "final kaf"}, IsSofit: true},
	'ם': {Char: 'ם', Name: "Mem Sofit", Meaning: "water", Position: 13, Aliases: []string{"mem sofit", "final mem"}, IsSofit: true},
	'ן': {Char: 'ן', Name: "Nun Sofit", Meaning: "fish", Position: 14, Aliases: []string{"nun sofit", "final nun"}, IsSofit: true},
	'ף': {Char: 'ף', Name: "Pe Sofit", Meaning: "mouth", Position: 17, Aliases: []string{"pe sofit", "peh sofit", "final pe"}, IsSofit: true},
	'ץ': {Char: 'ץ', Name: "Tsade Sofit", Meaning: "fish hook", Position: 18, Aliases: []string{"tsade sofit", "tzade sofit", "final tsade"}, IsSofit: true},
}
