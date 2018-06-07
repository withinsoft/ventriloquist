package proxytag

import (
	"testing"
	"testing/quick"
)

func TestFuzzParse(t *testing.T) {
	testParse := func(inp string) bool {
		_, err := Parse(inp, Nameslash, Sigils, HalfSigilStart, HalfSigilEnd)
		if err != nil && err != ErrNoMatch {
			t.Logf("error in parsing %q: %v", inp, err)
			return false
		}

		return true
	}

	err := quick.Check(testParse, &quick.Config{MaxCount: 65536 * 2})
	if err != nil {
		t.Fatal(err)
	}
}
