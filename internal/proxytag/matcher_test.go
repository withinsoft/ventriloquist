package proxytag

import "testing"

func TestMatcher(t *testing.T) {
	cases := []testCase{
		{
			err: ErrNoMatch,
		},
		{
			input: "Drake\\ You used to call me on my cellphone...",
			output: Match{
				Name:   "Drake",
				Method: "Nameslash",
				Body:   "You used to call me on my cellphone...",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = Nameslash
		t.Run(cs.input, cs.Run)
	}
}
