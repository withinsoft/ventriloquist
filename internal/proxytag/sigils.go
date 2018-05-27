package proxytag

import (
	"unicode"
  	"regexp"
)

// removes the first and last character of a string, analogous to
// shucking off the husk of an ear of corn.
func Shuck(victim string,firstlen int,lastlen int) string {
	return victim[firstlen: len(victim)-lastlen]
}

func isSigil(inp rune) bool {
	switch inp {
	case ';', '.', '?', '!':
		return false
	}

	return unicode.IsSymbol(inp) || unicode.IsPunct(inp)
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
  	var endRegex = regexp.MustCompile(`^[^\w\s]*`)
	lst := Reverse(endRegex.FindStringSubmatch(Reverse(message))[0])
	body := message[:len(message)-len(lst)]
	if len(lst) < 1 {
		return Match{}, ErrNoMatch
	}
	if !isSigil(rune(lst[0])) {
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
  	var startRegex = regexp.MustCompile(`^[^\w\s]*`)
	fst := startRegex.FindString(message)
	body := message[:len(fst)]
	if len(fst) < 1 {
		return Match{}, ErrNoMatch
	}
	if !isSigil(rune(fst[0])) {
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

	var startRegex = regexp.MustCompile(`^[^\w\s]*`)
	fst := startRegex.FindString(message)
	
	if len(fst) < 1 {
		return Match{}, ErrNoMatch
	}
	
	if !isSigil(rune(fst[0])) {
		return Match{}, ErrNoMatch
	}
	
	lst := Reverse(startRegex.FindStringSubmatch(Reverse(message))[0])
	
	if len(lst) < 1 {
		return Match{}, ErrNoMatch
	}
	
	if !isSigil(rune(lst[0])) {
		return Match{}, ErrNoMatch
	}
	
	body := Shuck(message,len(fst),len(lst))

	// prevent mistakes like `[ <@72838115944828928>` being mis-read
	if fst != '<' && lst == '>' {
		return Match{}, ErrNoMatch
	}


	return Match{
		InitialSigil: string(fst),
		EndSigil:     string(lst),
		Method:       "Sigils",
		Body:         body,
	}, nil
}
func Reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

