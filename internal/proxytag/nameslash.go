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

	fl := strings.Split(message, " ")
	f0 := fl[0]

	var cmp string
	// the backslash MUST be present in the first word
	if strings.Contains(f0, `\`) {
		cmp = `\`
	}
	if strings.Contains(f0, `:`) {
		cmp = ":"
	}

	if cmp == "" {
		return Match{}, ErrNoMatch
	}

	// the backslash MUST be the last character in the first word
	if string(f0[len(f0)-1]) != cmp {
		return Match{}, ErrNoMatch
	}

	name := f0[:len(f0)-1]
	body := strings.Join(fl[1:], " ")
	return Match{
		Name:   name,
		Method: "Nameslash",
		Body:   body,
	}, nil
}
