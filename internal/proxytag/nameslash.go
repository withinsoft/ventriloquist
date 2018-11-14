package proxytag

import (
	"errors"
	"strings"
)

// Nameslash errors
var (
	ErrSlashMustBeAtEnd = errors.New("proxytag: slash must be at the end of the string")
)

// Nameslash parses the "name-slash" method of proxy tagging.
//
// Given a message of the form:
//
//     Nicole\ hi there
//
// This returns:
//
//     Match{Name: "Nicole", Method: "Nameslash", Body: "hi there"}, nil
func Nameslash(message string) (Match, error) {
	if message == "" {
		return Match{}, ErrNoMatch
	}

	var cmp string

	for _, sigil := range []string{`\`, `:`, `/`, ">"} {
		if strings.Contains(message, sigil) {
			cmp = sigil
		}
	}

	if cmp == "" {
		return Match{}, ErrNoMatch
	}

	fl := strings.Split(message, cmp)
	f0 := fl[0]

	name := f0[:len(f0)]
	body := strings.TrimSpace(strings.Join(fl[1:], cmp))
	return Match{
		Name:   name,
		Method: "Nameslash",
		Body:   body,
	}, nil
}
