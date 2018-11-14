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

// Matchcerobj is a hack wrapper to return dynamically generated functions from a method
type Matcherobj struct {
	Matcher Matcher
}

// Parse parses a message with a list of matchers and returns the
func Parse(message string, matcherobjs ...Matcherobj) (Match, error) {
	if len(message) < 2 {
		return Match{}, ErrNoMatch
	}

	for _, matobj := range matcherobjs {
		mm, merr := matobj.Matcher(message)
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

//this replaces the existing matchers (in message decoding, not sysmate init) by taking a db match and creating a wrapper for a Matcher that checks just for that
func CreateMatcherobj(m Match) Matcherobj {
	//the message is already assumed trimmed for Matcher purposes
	switch method := m.Method; method {
		case "Nameslash":
			return &Matcherobj{
				Matcher: func(message string) (Match, error) {
					//check if name is there
					if strings.HasPrefix(message, m.Name) {
						message = strings.TrimSpace(message[len(m.Name):])
						//then check if one of the separators is there
						for _, sigil := range []string{`\`, `:`, `/`, '>'} {
							if string.HasPrefix(message, sigil) {
								//the rest trimmed is the message
								m.Body = strings.TrimSpace(message[len(sigil):])
								return m, nil
							}
						}
					}
					return m, ErrNoMatch						
				}
			}
		case "Sigils", "HalfSigilStart", "HalfSigilEnd":
			return &Matcherobj{
				Matcher: func(message string) (Match, error) {
					//check if there's the begin and end
					if strings.HasPrefix(message, m.InitialSigil) && strings.HasSuffix(message, m.EndSigil) {
						//the rest trimmed is the message
						m.Body = strings.TrimSpace(message[len(m.InitialSigil):len(message)-len(m.EndSigil)])
						return m, nil
					} 
					return m, ErrNoMatch
				}
			}
		default:
			// should never happen until new proxy methods are implemented
			return &Matcherobj{}
	}
}