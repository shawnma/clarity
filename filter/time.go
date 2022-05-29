package filter

import (
	"fmt"
	"strconv"
	"strings"
)

type TimeOfDay struct {
	Hour   int8
	Minute int8
	Second int8
}

func (t *TimeOfDay) Validate() error {
	if t.Hour < 0 || t.Hour > 23 {
		return fmt.Errorf("invalid hour for time of day: %d", t.Hour)
	}
	if t.Minute < 0 || t.Minute > 59 {
		return fmt.Errorf("invalid minute for time of day: %d", t.Minute)
	}
	if t.Second < 0 || t.Second > 59 {
		return fmt.Errorf("invalid second for time of day: %d", t.Second)
	}
	return nil
}

func (t *TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

func (t *TimeOfDay) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	p := strings.Split(s, ":")
	if len(p) < 2 || len(p) > 3 {
		return fmt.Errorf("invalid time of day, it should be in hh:mm or hh:mm:ss format: %s", s)
	}
	if h, err := strconv.Atoi(p[0]); err != nil {
		return fmt.Errorf("invalid hour: %s for %s", p[0], s)
	} else {
		t.Hour = int8(h)
	}
	if m, err := strconv.Atoi(p[1]); err != nil {
		return fmt.Errorf("invalid minute: %s for %s", p[1], s)
	} else {
		t.Minute = int8(m)
	}
	if len(p) == 3 {
		if ss, err := strconv.Atoi(p[2]); err != nil {
			return fmt.Errorf("invalid minute: %s for %s", p[2], s)
		} else {
			t.Second = int8(ss)
		}
	}
	err := t.Validate()
	if err != nil {
		return fmt.Errorf("invalid Day of time %s: %s", s, err)
	}
	return nil
}

func (t *TimeOfDay) isBefore(t2 *TimeOfDay) bool {
	return t.Hour < t2.Hour || (t.Hour == t2.Hour && t.Minute < t2.Minute) ||
		(t.Hour == t2.Hour && t.Minute == t2.Minute && t.Second < t2.Second)
}

type TimeRange struct {
	Begin TimeOfDay
	End   TimeOfDay
}
