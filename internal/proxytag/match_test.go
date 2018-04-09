package proxytag

import "testing"

type testCase struct {
	matcher Matcher
	input   string
	output  Match
	err     error
}

func (cs testCase) Run(t *testing.T) {
	m, err := Parse(cs.input, cs.matcher)

	if cs.output != m {
		t.Logf("expected: %v", cs.output)
		t.Logf("output:   %v", m)
		t.Error("match mismatch")
	}

	if cs.err == nil && err != nil {
		t.Fatalf("error found: %v", err)
	}

	if cs.err != nil {
		if cs.err != err {
			t.Errorf("wanted error %v, got: %v", cs.err, err)
		}
	}
}
