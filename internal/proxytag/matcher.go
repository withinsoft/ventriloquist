package proxytag

import (
	"encoding/json"
	"os/exec"
	"errors"
)

type Matcher struct {
	Prefix string
	Suffix string
	Systemmate string
}

type Match struct {
	Prefix string
	Suffix string
	Body string
	Systemmate string
}

type OldMatch struct {
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

func (m OldMatch) Matchers() []Matcher {
	print("METHOD: " + m.Method)
	matchers := make([]Matcher, 0)
	switch m.Method {
	case "Nameslash":
		for _, sep := range []string{"\\", ":", "/", ">"} {
			matchers = append(matchers, Matcher{
				Prefix: m.Name + sep,
				Suffix: "",
				Systemmate: m.Name,
			})
		}
	case "Sigils":
		matchers = append(matchers, Matcher{
			Prefix: m.InitialSigil,
			Suffix: m.EndSigil,
			Systemmate: m.Name,
		})
	case "HalfSigilStart":
		matchers = append(matchers, Matcher{
			Prefix: m.InitialSigil,
			Suffix: "",
			Systemmate: m.Name,
		})
	}
	return matchers
}

func (m Matcher) String() string {
	return m.Prefix + " text " + m.Suffix
}

func MatchMessage(msg string, matchers []Matcher) (Match, error) {
	requestMatchers := make([]map[string]interface{}, len(matchers))
	for i, matcher := range matchers {
		requestMatchers[i] = map[string]interface{}{
			"matcherPrefix": matcher.Prefix,
			"matcherSuffix": matcher.Suffix,
			"matcherSystemMate": matcher.Systemmate,
		}
	}

	request := map[string]interface{}{
		"tag": "RequestMatchMessage",
		"contents": map[string]interface{}{
			"messageBody": msg,
			"messageMatchers": requestMatchers,
		},
	}

	response, err := matcherExec(request)
	if err != nil {
		return Match{}, err
	}

	switch tag := response["tag"].(string); tag {
	case "ResponseError":
		return Match{}, errors.New(response["contents"].(string))
	case "ResponseMatch":
		contents := response["contents"].(map[string]interface{})
		return Match{
			Prefix: contents["matchPrefix"].(string),
			Suffix: contents["matchSuffix"].(string),
			Body: contents["matchBody"].(string),
			Systemmate: contents["matchSystemMate"].(string),
		}, nil
	default:
		return Match{}, errors.New("error: unexpected response from proxy-matcher")
	}
}

func DetectMatcher(msg string, systemmate string) (Matcher, error) {
	request := map[string]interface{}{
		"tag": "RequestDetectMatcher",
		"contents": map[string]interface{}{
			"detectMatcherMessage": msg,
			"detectMatcherSystemMate": systemmate,
		},
	}

	response, err := matcherExec(request)
	if err != nil {
		return Matcher{}, err
	}

	switch tag := response["tag"].(string); tag {
	case "ResponseError":
		return Matcher{}, errors.New(response["contents"].(string))
	case "ResponseMatcher":
		contents := response["contents"].(map[string]interface{})
		suffix := ""
		if contents["matcherSuffix"] != nil {
			suffix = contents["matcherSuffix"].(string)
		}
		return Matcher{
			Prefix: contents["matcherPrefix"].(string),
			Suffix: suffix,
			Systemmate: contents["matcherSystemMate"].(string),
		}, nil
	default:
		return Matcher{}, errors.New("error: unexpected response from proxy-matcher")
	}
}

func matcherExec(request map[string]interface{}) (map[string]interface{}, error) {
	cmd := exec.Command("proxy-matcher")
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	err = json.NewEncoder(cmdStdin).Encode(request)
	if err != nil {
		return nil, err
	}
	cmdStdin.Close()
	cmdStdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var response map[string]interface{}
	err = json.Unmarshal(cmdStdout, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
