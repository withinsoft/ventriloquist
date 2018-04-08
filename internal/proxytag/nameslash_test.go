package proxytag

import (
	"fmt"
	"testing"
)

func TestNameslash(t *testing.T) {
	cases := []struct {
		input string
		match Match
		err   error
	}{
		{
			err: ErrNoMatch,
		},
		{
			input: "foo bar",
			err:   ErrNoMatch,
		},
		{
			input: "\\ hi",
			err:   ErrNoNameGiven,
		},
		{
			input: "foo\\bar hi",
			err:   ErrNoMatch,
		},
		{
			input: "Nicole\\ hi there",
			match: Match{
				Name:   "Nicole",
				Method: "Nameslash",
				Body:   "hi there",
			},
		},
	}

	for _, cs := range cases {
		t.Run(fmt.Sprint(cs), func(t *testing.T) {
			m, err := Nameslash(cs.input)

			if cs.match != m {
				t.Logf("expected: %v", cs.match)
				t.Logf("output:   %v", m)
				t.Fatal("match mismatch")
			}

			if cs.err != err {
				t.Fatalf("wanted error %v, got: %v", cs.err, err)
			}
		})
	}
}
