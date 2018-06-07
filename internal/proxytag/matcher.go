package proxytag

import (
	"errors"
	"strings"
)

// Match is the result of a proxy tag scraping.
type Match struct {
	// Name is the name of the systemmate, if the proxy method supplies this
	Name string `json:"name,omitempty"`
	// IntialSigil and EndSigil are the beginning and end non-alphanumeric
	// text to signify the speaker
	InitialSigil string `json:"initial_sigil,omitempty"`
	EndSigil     string `json:"end_sigil,omitempty"`
	// Method is the proxy method the scraper is looking for.
	Method string `json:"method"`
	// Body is the rest of what the systemmate said.
	Body string `json:"body"`
}

func (m Match) String() string {
	sb := strings.Builder{}

	sb.WriteString("Method: " + m.Method)

	if m.Name != "" {
		sb.WriteString(", Name: " + m.Name)
	}

	if m.InitialSigil != "" {
		sb.WriteString(", Initial sigil: " + m.InitialSigil)
	}

	if m.EndSigil != "" {
		sb.WriteString(", End sigil: " + m.EndSigil)
	}

	return sb.String()
}

// Global errors.
var (
	ErrNoMatch = errors.New("proxytag: no match")
)

// Matcher is a function that can parse string data for proxied text.
//
// If no match is found, ErrNoMatch should be returned so the stack can continue
// processing.
type Matcher func(string) (Match, error)

// Parse parses a message with a list of matchers and returns the
func Parse(message string, matchers ...Matcher) (Match, error) {
	if len(message) < 3 {
		return Match{}, ErrNoMatch
	}

	for _, mat := range matchers {
		mm, merr := mat(message)
		if merr != nil {
			if merr == ErrNoMatch {
				continue
			}

			return mm, merr
		}

		return mm, nil
	}

	return Match{}, ErrNoMatch
}
