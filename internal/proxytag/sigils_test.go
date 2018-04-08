package proxytag

import "testing"

func TestShuck(t *testing.T) {
	s := Shuck("[memes]")
	if s != "memes" {
		t.Fatalf("wanted memes, got: %v", s)
	}
}

func TestHalfSigils(t *testing.T) {
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
				Method:       "HalfSigils",
				Body:         "memes",
			},
		},
	}

	for _, cs := range cases {
		cs.matcher = HalfSigils
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
