package util_test

import (
	"testing"

	"shawnma.com/clarity/util"
)

func TestSearch(t *testing.T) {
	cases := []struct {
		name     string
		data     []string
		input    string
		expValue bool
	}{
		{
			name:     "works",
			data:     []string{"com/google"},
			input:    "com/google/play",
			expValue: true,
		},
		{
			name:     "nonmatch",
			data:     []string{"com/google"},
			input:    "com/hehe",
			expValue: false,
		},
		{
			name:     "empty input",
			data:     []string{"com/google"},
			input:    "",
			expValue: false,
		},
		{
			name:     "match longer path",
			data:     []string{"com/google", "com/google/play", "com/google/play/data"},
			input:    "com/google/play/data",
			expValue: true,
		},
	}

	for _, c := range cases {
		t.Logf("Testing %s", c.name)
		var trie util.PathTrie[bool]
		for _, d := range c.data {
			trie.Put(d, true)
		}
		v := trie.Search(c.input)
		if v != c.expValue {
			t.Fatalf("Wrong value output, expected: %t, got: %t", c.expValue, v)
		}
	}
}
