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
	Result  bool
	Message string
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
	if minutes > 60 {
		return setResult{false, "You can't set the time greater than 1 hour"}
	}
	t := time.Now()
	d := time.Duration(minutes) * time.Minute
	log.Printf("ADD DURATION %s %s", host, d)
	t = t.Add(d)
	e.ExpireTime = &t
	return e
}
