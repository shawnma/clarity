package filter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (h *Filter) HttpHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Println(req.URL.Path)
		var result any
		switch req.URL.Path {
		case "/config/settings":
			result = h.getSettings()
		case "/config/blocked":
			result = h.getBlockedInfo(req)
		case "/config/set":
			result = h.setTemp(req)
		}
		json.NewEncoder(w).Encode(result)
	}
	return http.HandlerFunc(fn)
}

func (f *Filter) getSettings() any {
	return f.tree.Values()
}

func (f *Filter) getBlockedInfo(req *http.Request) any {
	id, err := strconv.Atoi(req.URL.Query().Get("id"))
	if err != nil {
		return setResult{false, "Wrong id"}
	}
	e := f.findEntry(id)
	if e == nil {
		return setResult{false, fmt.Sprintf("Config id %d not found", id)}
	}
	return e
}

type setResult struct {
	Result  bool
	Message string
}

func (f *Filter) findEntry(id int) *Entry {

	v := f.tree.Values()
	if v == nil {
		return nil
	}

	// should have a map ...
	for _, e := range v {
		if e.Id == id {
			return e
		}
	}
	return nil
}

func (f *Filter) setTemp(req *http.Request) any {
	id, err := strconv.Atoi(req.URL.Query().Get("id"))
	if err != nil {
		return setResult{false, "Wrong id"}
	}

	minutes, err := strconv.Atoi(req.URL.Query().Get("t"))
	if err != nil {
		return setResult{false, "Minutes is not an interger: " + err.Error()}
	}
	if minutes > 60 {
		return setResult{false, "You can't set the time greater than 1 hour"}
	}
	t := time.Now()
	d := time.Duration(minutes) * time.Minute

	e := f.findEntry(id)
	if e == nil {
		return setResult{false, fmt.Sprintf("Config id %d not found", id)}
	}
	log.Printf("ADD DURATION %s %s", e.Policy.Path, d)
	t = t.Add(d)
	e.ExpireTime = &t
	return e
}
