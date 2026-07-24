package gematria

// hechrachi holds the standard (mispar hechrachi) values for all 27 Hebrew runes.
// Sofit forms carry the same value as their normal form.
var hechrachi = map[rune]int{
	'א': 1, 'ב': 2, 'ג': 3, 'ד': 4, 'ה': 5,
	'ו': 6, 'ז': 7, 'ח': 8, 'ט': 9, 'י': 10,
	'כ': 20, 'ל': 30, 'מ': 40, 'נ': 50, 'ס': 60,
	'ע': 70, 'פ': 80, 'צ': 90, 'ק': 100, 'ר': 200,
	'ש': 300, 'ת': 400,
	// sofit — same value as normal form
	'ך': 20, 'ם': 40, 'ן': 50, 'ף': 80, 'ץ': 90,
}

// gadol holds the mispar gadol (great) values. Standard letters are identical to
// hechrachi; sofit forms receive extended values (500–900).
var gadol = map[rune]int{
	'א': 1, 'ב': 2, 'ג': 3, 'ד': 4, 'ה': 5,
	'ו': 6, 'ז': 7, 'ח': 8, 'ט': 9, 'י': 10,
	'כ': 20, 'ל': 30, 'מ': 40, 'נ': 50, 'ס': 60,
	'ע': 70, 'פ': 80, 'צ': 90, 'ק': 100, 'ר': 200,
	'ש': 300, 'ת': 400,
	// sofit — extended values
	'ך': 500, 'ם': 600, 'ן': 700, 'ף': 800, 'ץ': 900,
}

// siduri holds the mispar siduri (ordinal) values: each letter is numbered 1–22
// by its position. Sofit forms share the ordinal of their normal form.
var siduri = map[rune]int{
	'א': 1, 'ב': 2, 'ג': 3, 'ד': 4, 'ה': 5,
	'ו': 6, 'ז': 7, 'ח': 8, 'ט': 9, 'י': 10,
	'כ': 11, 'ל': 12, 'מ': 13, 'נ': 14, 'ס': 15,
	'ע': 16, 'פ': 17, 'צ': 18, 'ק': 19, 'ר': 20,
	'ש': 21, 'ת': 22,
	// sofit — same ordinal as normal form
	'ך': 11, 'ם': 13, 'ן': 14, 'ף': 17, 'ץ': 18,
}

// atbash holds the mispar atbash (cipher) values. Each letter takes the
// hechrachi value of its mirror: position 1 ↔ 22, 2 ↔ 21, …, 11 ↔ 12.
// Sofit forms mirror through their normal form's pair.
var atbash = map[rune]int{
	'א': 400, 'ב': 300, 'ג': 200, 'ד': 100, 'ה': 90,
	'ו': 80, 'ז': 70, 'ח': 60, 'ט': 50, 'י': 40,
	'כ': 30, 'ל': 20, 'מ': 10, 'נ': 9, 'ס': 8,
	'ע': 7, 'פ': 6, 'צ': 5, 'ק': 4, 'ר': 3,
	'ש': 2, 'ת': 1,
	// sofit — mirrors through normal form's pair
	'ך': 30, 'ם': 10, 'ן': 9, 'ף': 6, 'ץ': 5,
}

// systemValues aggregates all four tables for O(1) dispatch by System constant.
var systemValues = map[System]map[rune]int{
	Hechrachi: hechrachi,
	Gadol:     gadol,
	Siduri:    siduri,
	Atbash:    atbash,
}

// atbashMirror maps each rune to its Atbash partner rune.
// The mapping is bidirectional: position 1↔22, 2↔21, …, 11↔12.
// Sofit forms mirror to the normal form's pair rune.
var atbashMirror = map[rune]rune{
	'א': 'ת', 'ב': 'ש', 'ג': 'ר', 'ד': 'ק', 'ה': 'צ',
	'ו': 'פ', 'ז': 'ע', 'ח': 'ס', 'ט': 'נ', 'י': 'מ',
	'כ': 'ל', 'ל': 'כ', 'מ': 'י', 'נ': 'ט', 'ס': 'ח',
	'ע': 'ז', 'פ': 'ו', 'צ': 'ה', 'ק': 'ד', 'ר': 'ג',
	'ש': 'ב', 'ת': 'א',
	// sofit forms substitute to the normal form's pair
	'ך': 'ל', 'ם': 'י', 'ן': 'ט', 'ף': 'ו', 'ץ': 'ה',
}

// ValidSystems returns all four gematria systems in a stable order.
// Used in error messages and flag validation.
func ValidSystems() []System {
	return []System{Hechrachi, Gadol, Siduri, Atbash}
}
