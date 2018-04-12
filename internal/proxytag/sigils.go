package proxytag

import (
	"unicode"
)

// Shuck removes the first and last character of a string, analogous to
// shucking off the husk of an ear of corn.
func Shuck(victim string) string {
	return victim[1 : len(victim)-1]
}

func isSigil(inp rune) bool {
	return unicode.IsSymbol(inp) || unicode.IsPunct(inp)
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

	lst := rune(message[len(message)-1])
	body := message[:len(message)-1]
	if !isSigil(lst) {
		return Match{}, ErrNoMatch
	}

	return Match{
		EndSigil: string(lst),
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

	fst := rune(message[0])
	body := message[1:]
	if !isSigil(fst) {
		return Match{}, ErrNoMatch
	}

	return Match{
		InitialSigil: string(fst),
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

	fst := rune(message[0])
	lst := rune(message[len(message)-1])
	body := Shuck(message)

	if !isSigil(fst) {
		return Match{}, ErrNoMatch
	}

	if !isSigil(lst) {
		return Match{}, ErrNoMatch
	}

	return Match{
		InitialSigil: string(fst),
		EndSigil:     string(lst),
		Method:       "Sigils",
		Body:         body,
	}, nil
}
