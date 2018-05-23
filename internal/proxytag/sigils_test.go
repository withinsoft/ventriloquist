package proxytag

import "testing"

func TestShuck(t *testing.T) {
	s := Shuck("[memes]")
	if s != "memes" {
		t.Fatalf("wanted memes, got: %v", s)
	}
}

func TestIsSigil(t *testing.T) {
	cases := []rune{
		'[',
		'$',
	}

	for _, cs := range cases {
		t.Run(string(cs), func(t *testing.T) {
			if !isSigil(cs) {
				t.Fatalf("not sigil: %s", string(cs))
			}
		})
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
		{
			input: "$memes",
			output: Match{
				InitialSigil: "$",
				Method:       "HalfSigilStart",
				Body:         "memes",
			},
		},
		{
			input: "[ <@72838115944828928> test",
			output: Match{
				InitialSigil: "[",
				Method: "HalfSigilStart",
				Body: " <@72838115944828928> test",
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
		{
			input: "memes$",
			output: Match{
				EndSigil: "$",
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
		{
			input: "$memes$",
			output: Match{
				InitialSigil: "$",
				EndSigil:     "$",
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
