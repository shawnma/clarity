package filter

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestTimeDay(t *testing.T) {
	d := TimeOfDay{1, 2, 3}
	r, _ := json.Marshal(d)
	fmt.Printf("%s\n", string(r))
}

func TestTimeRange(t *testing.T) {
	b := TimeOfDay{1, 2, 3}
	e := TimeOfDay{4, 5, 6}
	tx := TimeRange{b, e}
	r, _ := json.Marshal(tx)
	fmt.Printf("%s\n", string(r))
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		expected *TimeOfDay
		err      string
	}{{
		"1000",
		nil,
		"invalid",
	}, {
		"10:",
		nil,
		"invalid",
	}, {
		"a:b",
		nil,
		"invalid",
	}, {
		"4:5",
		&TimeOfDay{4, 5, 0},
		"",
	}, {
		"4:5:6",
		&TimeOfDay{4, 5, 6},
		"",
	}, {
		"3:4:5:5:9",
		nil,
		"invalid",
	}, {
		"55:9",
		nil,
		"invalid",
	},
	}
	for _, tc := range tests {
		var tt TimeOfDay
		err := json.Unmarshal([]byte("\""+tc.input+"\""), &tt)
		fmt.Printf("input: %s err: %s, got_err: %s, got: %s\n", tc.input, tc.err, err, tt.String())
		if tc.err != "" && !strings.Contains(err.Error(), tc.err) {
			t.Errorf("Expected error %s, got: %s", tc.err, err.Error())
		}
		if tc.expected != nil && *tc.expected != tt {
			t.Errorf("Expected result %s, got: %s", tc.expected.String(), tt.String())
		}
	}
}
