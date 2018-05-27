package proxytag

import "testing"

func TestShuck(t *testing.T) {
	s := Shuck("[memes]")
	if s != "memes" {
		t.Fatalf("wanted memes, got: %v", s)
	}
}

func TestLastRune(t *testing.T) {
	cases := []struct {
		inp string
		out rune
	}{
		{}, // nothing should be nothing
		{
			inp: "hi",
			out: 'i',
		},
		{
			inp: "你好",
			out: '好',
		},
	}

	for _, cs := range cases {
		t.Run(cs.inp, func(t *testing.T) {
			result := lastRune(cs.inp)
			if result != cs.out {
				t.Fatalf("wanted: %s, got: %s", string(cs.out), string(result))
			}
		})
	}
}

func TestFirstRune(t *testing.T) {
	cases := []struct {
		inp string
		out rune
	}{
		{}, // nothing should be nothing
		{
			inp: "hi",
			out: 'h',
		},
		{
			inp: "你好",
			out: '你',
		},
	}

	for _, cs := range cases {
		t.Run(cs.inp, func(t *testing.T) {
			result := firstRune(cs.inp)
			if result != cs.out {
				t.Fatalf("wanted: %s, got: %s", string(cs.out), string(result))
			}
		})
	}
}

func TestIsSigil(t *testing.T) {
	cases := []struct {
		inp  rune
		good bool
	}{
		{
			inp:  '[',
			good: true,
		},
		{
			inp:  '$',
			good: true,
		},
		{
			inp:  ';',
			good: false,
		},
	}

	for _, cs := range cases {
		t.Run(string(cs.inp), func(t *testing.T) {
			if result := isSigil(cs.inp); result != cs.good {
				t.Fatalf("wanted %v for %s, got: %v", cs.good, string(cs.inp), result)
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
			input: "[[memes",
			output: Match{
				InitialSigil: "[[",
				Method: "HalfSigilStart",
				Body: "memes",
			},
		},
		{
			input: "[ <@72838115944828928> test",
			output: Match{
				InitialSigil: "[",
				Method:       "HalfSigilStart",
				Body:         " <@72838115944828928> test",
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
		{
			input: "memes]]",
			output: Match{
				EndSigil: "]]",
				Method: "HalfSigilEnd",
				Body: "memes",
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
		{
			input: "[ <@72838115944828928>",
			err:   ErrNoMatch,
		},
	}

	for _, cs := range cases {
		cs.matcher = Sigils
		t.Run(cs.input, cs.Run)
	}
}
