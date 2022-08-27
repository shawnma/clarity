package util

import (
	"testing"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		name     string
		data     []string
		input    string
		expValue bool
	}{
		{
			name:     "works",
			data:     []string{"google.com/log"},
			input:    "google.com",
			expValue: false,
		},
		{
			name:     "nonmatch",
			data:     []string{"google.com"},
			input:    "com",
			expValue: false,
		},
		{
			name:     "empty input",
			data:     []string{"google.com"},
			input:    "",
			expValue: false,
		},
		{
			name:     "domain only",
			data:     []string{"google.com"},
			input:    "google.com",
			expValue: true,
		},
		{
			name:     "sub domain only",
			data:     []string{"google.com"},
			input:    "player.google.com",
			expValue: true,
		},
		{
			name:     "match longer path",
			data:     []string{"google.com", "google.com/data"},
			input:    "play.google.com/data/log",
			expValue: true,
		},
		{
			name:     "complex case",
			data:     []string{"google.com", "google.com/data", "play.google.com/blah"},
			input:    "play.google.com/data/log",
			expValue: true,
		},
		{
			name:     "mixed case",
			data:     []string{"google.com", "google.com/data", "play.google.com/blah"},
			input:    "google.com",
			expValue: true,
		},
	}

	for _, c := range cases {
		t.Logf("Testing %s", c.name)
		var matcher UrlMatch[bool]
		for _, d := range c.data {
			matcher.Add(d, true)
		}
		h, p := splitUrl(c.input)
		v := matcher.Match(h, p)
		if v != c.expValue {
			t.Errorf("Wrong value output, expected: %t, got: %t", c.expValue, v)
		}
	}
}
