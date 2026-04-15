package gematria

import (
	"errors"
	"strings"
	"testing"
)

// InvalidCharError

func TestInvalidCharErrorMessage(t *testing.T) {
	e := &InvalidCharError{Char: 'X', Position: 3, Input: "abXde"}
	msg := e.Error()
	if !strings.Contains(msg, "X") {
		t.Errorf("Error() = %q: want invalid char 'X' in message", msg)
	}
	if !strings.Contains(msg, "3") {
		t.Errorf("Error() = %q: want position 3 in message", msg)
	}
}

func TestInvalidCharErrorFields(t *testing.T) {
	e := &InvalidCharError{Char: 'Z', Position: 7, Input: "hello Z"}
	if e.Char != 'Z' {
		t.Errorf("Char = %c, want Z", e.Char)
	}
	if e.Position != 7 {
		t.Errorf("Position = %d, want 7", e.Position)
	}
	if e.Input != "hello Z" {
		t.Errorf("Input = %q, want %q", e.Input, "hello Z")
	}
}

func TestInvalidCharErrorImplementsError(t *testing.T) {
	var e error = &InvalidCharError{Char: 'X', Position: 0, Input: "X"}
	if e.Error() == "" {
		t.Error("Error() returned empty string")
	}
	var target *InvalidCharError
	if !errors.As(e, &target) {
		t.Error("errors.As failed for *InvalidCharError")
	}
}

// UnknownNameError

func TestUnknownNameErrorNoSuggestions(t *testing.T) {
	e := &UnknownNameError{Name: "bogus", Position: 0, Suggestions: nil}
	msg := e.Error()
	if !strings.Contains(msg, "bogus") {
		t.Errorf("Error() = %q: want name %q in message", msg, "bogus")
	}
	if strings.Contains(msg, "did you mean") {
		t.Errorf("Error() = %q: should not contain 'did you mean' when Suggestions is nil", msg)
	}
}

func TestUnknownNameErrorWithSuggestions(t *testing.T) {
	e := &UnknownNameError{Name: "alph", Position: 0, Suggestions: []string{"aleph", "alef"}}
	msg := e.Error()
	if !strings.Contains(msg, "alph") {
		t.Errorf("Error() = %q: want name %q in message", msg, "alph")
	}
	if !strings.Contains(msg, "did you mean") {
		t.Errorf("Error() = %q: want 'did you mean' when suggestions present", msg)
	}
	if !strings.Contains(msg, "aleph") {
		t.Errorf("Error() = %q: want suggestion %q in message", msg, "aleph")
	}
}

func TestUnknownNameErrorFields(t *testing.T) {
	e := &UnknownNameError{Name: "foo", Position: 2, Suggestions: []string{"bar"}}
	if e.Name != "foo" {
		t.Errorf("Name = %q, want %q", e.Name, "foo")
	}
	if e.Position != 2 {
		t.Errorf("Position = %d, want 2", e.Position)
	}
	if len(e.Suggestions) != 1 || e.Suggestions[0] != "bar" {
		t.Errorf("Suggestions = %v, want [bar]", e.Suggestions)
	}
}

func TestUnknownNameErrorImplementsError(t *testing.T) {
	var e error = &UnknownNameError{Name: "x", Position: 0}
	var target *UnknownNameError
	if !errors.As(e, &target) {
		t.Error("errors.As failed for *UnknownNameError")
	}
}

// InvalidSystemError

func TestInvalidSystemErrorMessage(t *testing.T) {
	e := &InvalidSystemError{Name: "mystery", Valid: []System{Hechrachi, Gadol}}
	msg := e.Error()
	if !strings.Contains(msg, "mystery") {
		t.Errorf("Error() = %q: want name %q in message", msg, "mystery")
	}
	if !strings.Contains(msg, "hechrachi") {
		t.Errorf("Error() = %q: want valid system %q listed", msg, "hechrachi")
	}
	if !strings.Contains(msg, "gadol") {
		t.Errorf("Error() = %q: want valid system %q listed", msg, "gadol")
	}
}

func TestInvalidSystemErrorFields(t *testing.T) {
	valid := []System{Hechrachi, Gadol, Siduri, Atbash}
	e := &InvalidSystemError{Name: "bad", Valid: valid}
	if e.Name != "bad" {
		t.Errorf("Name = %q, want %q", e.Name, "bad")
	}
	if len(e.Valid) != 4 {
		t.Errorf("len(Valid) = %d, want 4", len(e.Valid))
	}
}

func TestInvalidSystemErrorImplementsError(t *testing.T) {
	var e error = &InvalidSystemError{Name: "bad", Valid: ValidSystems()}
	var target *InvalidSystemError
	if !errors.As(e, &target) {
		t.Error("errors.As failed for *InvalidSystemError")
	}
}

// UnknownWordError

func TestUnknownWordErrorNoSuggestions(t *testing.T) {
	e := &UnknownWordError{Input: "qzxw", Scheme: SchemeAcademic, Position: 0, Suggestions: nil}
	msg := e.Error()
	if !strings.Contains(msg, "qzxw") {
		t.Errorf("Error() = %q: want input %q in message", msg, "qzxw")
	}
	if !strings.Contains(msg, "academic") {
		t.Errorf("Error() = %q: want scheme %q in message", msg, "academic")
	}
	if strings.Contains(msg, "did you mean") {
		t.Errorf("Error() = %q: should not contain 'did you mean' when Suggestions is nil", msg)
	}
}

func TestUnknownWordErrorWithSuggestions(t *testing.T) {
	e := &UnknownWordError{Input: "shaloom", Scheme: SchemeAcademic, Position: 0, Suggestions: []string{"shalom"}}
	msg := e.Error()
	if !strings.Contains(msg, "shaloom") {
		t.Errorf("Error() = %q: want input %q in message", msg, "shaloom")
	}
	if !strings.Contains(msg, "did you mean") {
		t.Errorf("Error() = %q: want 'did you mean' when suggestions present", msg)
	}
	if !strings.Contains(msg, "shalom") {
		t.Errorf("Error() = %q: want suggestion %q in message", msg, "shalom")
	}
}

func TestUnknownWordErrorFields(t *testing.T) {
	e := &UnknownWordError{Input: "foo", Scheme: SchemeIsraeli, Position: 2, Suggestions: []string{"bar"}}
	if e.Input != "foo" {
		t.Errorf("Input = %q, want %q", e.Input, "foo")
	}
	if e.Scheme != SchemeIsraeli {
		t.Errorf("Scheme = %q, want %q", e.Scheme, SchemeIsraeli)
	}
	if e.Position != 2 {
		t.Errorf("Position = %d, want 2", e.Position)
	}
	if len(e.Suggestions) != 1 || e.Suggestions[0] != "bar" {
		t.Errorf("Suggestions = %v, want [bar]", e.Suggestions)
	}
}

func TestUnknownWordErrorImplementsError(t *testing.T) {
	var e error = &UnknownWordError{Input: "x", Scheme: SchemeAcademic, Position: 0}
	var target *UnknownWordError
	if !errors.As(e, &target) {
		t.Error("errors.As failed for *UnknownWordError")
	}
}

func TestUnknownWordErrorMessageFormat(t *testing.T) {
	// Message must be lowercase and follow the spec format.
	e := &UnknownWordError{Input: "qzxw", Scheme: SchemeAcademic, Position: 0}
	msg := e.Error()
	if msg[0] >= 'A' && msg[0] <= 'Z' {
		t.Errorf("Error() = %q: message must be lowercase (first char is uppercase)", msg)
	}
}

// InvalidSchemeError

func TestInvalidSchemeErrorMessage(t *testing.T) {
	e := &InvalidSchemeError{Name: "bogus", Valid: []Scheme{SchemeAcademic, SchemeIsraeli}}
	msg := e.Error()
	if !strings.Contains(msg, "bogus") {
		t.Errorf("Error() = %q: want name %q in message", msg, "bogus")
	}
	if !strings.Contains(msg, "academic") {
		t.Errorf("Error() = %q: want valid scheme %q listed", msg, "academic")
	}
	if !strings.Contains(msg, "israeli") {
		t.Errorf("Error() = %q: want valid scheme %q listed", msg, "israeli")
	}
}

func TestInvalidSchemeErrorFields(t *testing.T) {
	valid := []Scheme{SchemeAcademic, SchemeIsraeli}
	e := &InvalidSchemeError{Name: "nope", Valid: valid}
	if e.Name != "nope" {
		t.Errorf("Name = %q, want %q", e.Name, "nope")
	}
	if len(e.Valid) != 2 {
		t.Errorf("len(Valid) = %d, want 2", len(e.Valid))
	}
}

func TestInvalidSchemeErrorImplementsError(t *testing.T) {
	var e error = &InvalidSchemeError{Name: "bad", Valid: ValidSchemes()}
	var target *InvalidSchemeError
	if !errors.As(e, &target) {
		t.Error("errors.As failed for *InvalidSchemeError")
	}
}

func TestInvalidSchemeErrorMessageFormat(t *testing.T) {
	e := &InvalidSchemeError{Name: "bogus", Valid: ValidSchemes()}
	msg := e.Error()
	if msg[0] >= 'A' && msg[0] <= 'Z' {
		t.Errorf("Error() = %q: message must be lowercase", msg)
	}
}
