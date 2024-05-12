package config

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	r, _ := NewTimeRange("10:00", "11:00")
	r2, _ := NewTimeRange("21:00", "23:00")
	p := Policy{
		Path:         "youtube.com/chat",
		AllowedRange: []TimeRange{*r, *r2},
		MaxAllowed:   2 * time.Hour,
	}
	o, _ := yaml.Marshal(p)
	fmt.Println(string(o))
	var out Policy
	yaml.Unmarshal(o, &out)
	fmt.Printf("%+v", &out)
	if *r != out.AllowedRange[0] || *r2 != out.AllowedRange[1] {
		t.Errorf("Serialization problem")
	}
}
