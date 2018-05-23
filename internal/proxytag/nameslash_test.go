package proxytag

import (
	"fmt"
	"testing"
)

func TestNameslash(t *testing.T) {
	cases := []testCase{
		{
			err: ErrNoMatch,
		},
		{
			input: "foo bar",
			err:   ErrNoMatch,
		},
		{
			input: "foo\\bar hi",
			err:   ErrNoMatch,
		},
		{
			input: "Nicole\\ hi there",
			output: Match{
				Name:   "Nicole",
				Method: "Nameslash",
				Body:   "hi there",
			},
		},
		{
			input: "Nicole: hi there",
			output: Match{
				Name:   "Nicole",
				Method: "Nameslash",
				Body:   "hi there",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = Nameslash
		t.Run(fmt.Sprint(cs), cs.Run)
	}
}
