package proxytag

import "testing"

func TestShuck(t *testing.T) {
	s := Shuck("[memes]")
	if s != "memes" {
		t.Fatalf("wanted memes, got: %v", s)
	}
}

func TestHalfSigilStart(t *testing.T) {
	cases := []testCase{
		{
			input: "a",
			err:   ErrNoMatch,
		},
		{
			input: "aad",
			err:   ErrNoMatch,
		},
		{
			input: "[memes",
			output: Match{
				InitialSigil: "[",
				Method:       "HalfSigilStart",
				Body:         "memes",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = HalfSigilStart
		t.Run(cs.input, cs.Run)
	}
}

func TestHalfSigilEnd(t *testing.T) {
	cases := []testCase{
		{
			input: "a",
			err:   ErrNoMatch,
		},
		{
			input: "aad",
			err:   ErrNoMatch,
		},
		{
			input: "memes]",
			output: Match{
				EndSigil: "]",
				Method:   "HalfSigilEnd",
				Body:     "memes",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = HalfSigilEnd
		t.Run(cs.input, cs.Run)
	}
}

func TestSigls(t *testing.T) {
	cases := []testCase{
		{
			input: "as",
			err:   ErrNoMatch,
		},
		{
			input: "fast don't lie",
			err:   ErrNoMatch,
		},
		{
			input: "[as",
			err:   ErrNoMatch,
		},
		{
			input: "[memes]",
			output: Match{
				InitialSigil: "[",
				EndSigil:     "]",
				Method:       "Sigils",
				Body:         "memes",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = Sigils
		t.Run(cs.input, cs.Run)
	}
}
