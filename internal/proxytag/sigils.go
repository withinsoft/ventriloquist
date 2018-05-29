package proxytag

import (
	"unicode"
)

// removes the first and last character of a string, analogous to
// shucking off the husk of an ear of corn.
func Shuck(victim string, firstlen int, lastlen int) string {
	return victim[firstlen : len(victim)-lastlen]
}

func isSigil(inp rune) bool {
	switch inp {
	// english formatting characters
	case ';', '.', '?', '!', ',', '-':
		return false
	// discord formatting characters
	case '*', '_', '<', '>', '~':
		return false
	}

	return unicode.IsSymbol(inp) || unicode.IsPunct(inp)
}

func leadSigils(inp string) (string, string) {
	var sigils []rune
	var result []rune
	caught := false
	for _, r := range inp {
		if !caught && isSigil(r) {
			sigils = append(sigils, r)
		} else {
			caught = true
			result = append(result, r)
		}
	}

	return string(sigils), string(result)
}

func tailSigils(inp string) (string, string) {
	var sigils []rune
	var result []rune
	caught := false
	for _, r := range Reverse(inp) {
		if !caught && isSigil(r) {
			sigils = append(sigils, r)
		} else {
			caught = true
			result = append(result, r)
		}
	}

	return Reverse(string(sigils)), Reverse(string(result))
}

func firstRune(inp string) rune {
	for _, rn := range inp {
		return rn
	}

	return rune(0)
}

func lastRune(inp string) rune {
	var result rune
	for _, rn := range inp {
		result = rn
	}

	return result
}

// HalfSigilStart parses the "half sigil at the start" method of proxy tagging.
//
// Given a message of the form:
//
//     foo]
//
// This returns
//
//     Match{EndSigil:"]", Method: "HalfSigilEnd", Body: "foo"}
func HalfSigilEnd(message string) (Match, error) {
	if len(message) < 2 {
		return Match{}, ErrNoMatch
	}

	sigils, body := tailSigils(message)
	if sigils == "" {
		return Match{}, ErrNoMatch
	}

	return Match{
		EndSigil: sigils,
		Method:   "HalfSigilEnd",
		Body:     body,
	}, nil
}

// HalfSigilStart parses the "half sigil at the start" method of proxy tagging.
//
// Given a message of the form:
//
//     [foo
//
// This returns
//
//     Match{InitialSigil:"[", Method: "HalfSigils", Body: "foo"}
func HalfSigilStart(message string) (Match, error) {
	if len(message) < 2 {
		return Match{}, ErrNoMatch
	}

	sigils, body := leadSigils(message)
	if sigils == "" {
		return Match{}, ErrNoMatch
	}

	return Match{
		InitialSigil: sigils,
		Method:       "HalfSigilStart",
		Body:         body,
	}, nil
}

// Sigils parses the "sigils" method of proxy tagging.
//
// Given a message of the form:
//
//     [foo]
//
// This returns
//
//     Match{InitialSigil:"[", EndSigil: "]", Method: "Sigils", Body: "foo"}
func Sigils(message string) (Match, error) {
	if len(message) < 3 {
		return Match{}, ErrNoMatch
	}

	startSigils, body1 := leadSigils(message)
	endSigils, body := tailSigils(body1)

	if startSigils == "" || endSigils == "" {
		return Match{}, ErrNoMatch
	}

	return Match{
		InitialSigil: startSigils,
		EndSigil:     endSigils,
		Method:       "Sigils",
		Body:         body,
	}, nil
}

// Reverse reverses a string rune by rune.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
