//+build js

package main

import (
	"fmt"

	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/withinsoft/ventriloquist/internal/proxytag"
)

func main() {
	vecty.SetTitle("Ventriloquist Parsing Tester")
	vecty.RenderBody(&PageView{
		Input: "[I am listening for a sound beyond sound that stalks the nightland of my dreams]",
	})
}

// PageView is our main page component that takes input of proxied text and
// shows the parsed version of it.
type PageView struct {
	vecty.Core
	Input string
}

// Render implements the vecty.Component interface.
func (p *PageView) Render() vecty.ComponentOrHTML {
	mi := &MatchInfo{Input: p.Input}

	return elem.Body(
		// Display a textarea on the right-hand side of the page.
		elem.Div(
			vecty.Markup(
				vecty.Style("float", "right"),
			),
			elem.TextArea(
				vecty.Markup(
					vecty.Style("font-family", "monospace"),
					vecty.Property("rows", 14),
					vecty.Property("cols", 70),

					// When input is typed into the textarea, update the local
					// component state and rerender.
					event.Input(func(e *vecty.Event) {
						val := e.Target.Get("value").String()
						fmt.Println(val)
						mi.Input = val
						vecty.Rerender(mi)
					}),
				),
				vecty.Text(p.Input), // initial textarea text.
			),
		),

		mi,
	)
}

// MatchInfo shows match information based on the results of the proxytag parsing.
type MatchInfo struct {
	vecty.Core
	Input string
}

// Render shows the match information as a
func (m *MatchInfo) Render() vecty.ComponentOrHTML {
	match, err := proxytag.Parse(m.Input, proxytag.Nameslash, proxytag.Sigils, proxytag.HalfSigilEnd, proxytag.HalfSigilStart)
	if err != nil {
		return elem.Div(
			elem.Paragraph(vecty.Text("There was an error: " + err.Error())),
		)
	}

	str := `Match Results:
Name: %s
Intial Sigil: %s
End Sigil: %s
Method: %s
Body: %s`

	return elem.Div(
		elem.Code(
			elem.Preformatted(vecty.Text(fmt.Sprintf(str, match.Name, match.InitialSigil, match.EndSigil, match.Method, match.Body))),
		),
	)
}
