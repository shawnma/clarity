package filter

import (
	"encoding/json"
	"fmt"
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
	var e []*Entry
	f.t.Walk(func(key string, value *Entry) error {
		e = append(e, value)
		return nil
	})
	return e
}

func (f *Filter) getBlockedInfo(req *http.Request) any {
	e := f.t.Search(req.URL.Query().Get("s"))
	return e
}

type setResult struct {
	result  bool
	message string
}

func (f *Filter) setTemp(req *http.Request) any {
	host := req.URL.Query().Get("s")
	e := f.t.Search(host)
	if e == nil {
		return setResult{false, "Couldn't find the config for " + host}
	}
	fmt.Println(e)
	minutes, err := strconv.Atoi(req.URL.Query().Get("t"))
	if err != nil {
		return setResult{false, "Minutes is not an interger: " + err.Error()}
	}
	t := time.Now()
	t.Add(time.Duration(minutes) * time.Minute)
	e.ExpireTime = &t
	return e
}
