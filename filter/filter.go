package filter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/martian/v3"
	"shawnma.com/clarity/trie"
)

type Entry struct {
	ExpireTime time.Time
	Path       string // Path for displaying purpose
}

type Filter struct {
	t *trie.PathTrie[*Entry]
}

func NewFilter() *Filter {
	f := &Filter{}
	f.t = trie.NewPathTrie[*Entry]()
	f.t.Put("youtube.com", &Entry{
		ExpireTime: time.Now().Add(time.Hour),
		Path:       "youtube.com",
	})
	return f
}

func (h *Filter) HttpHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			var e []*Entry
			h.t.Walk(func(key string, value *Entry) error {
				e = append(e, value)
				return nil
			})
			json.NewEncoder(w).Encode(e)
		}
	}
	return http.HandlerFunc(fn)
}

// ModifyRequest return 403 if an entry is matched
func (f *Filter) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	url := req.URL
	path := url.Hostname() + "/" + url.RequestURI()
	entry := f.t.Search(path)
	if entry != nil && req.Method != "CONNECT" && entry.ExpireTime.After(time.Now()) {
		_, w, err := ctx.Session().Hijack()
		if err != nil {
			return err
		}
		resp := "HTTP/1.1 302 moved\nLocation: https://clarity.proxy/filter\nConnection: Close\n\n"
		w.Write([]byte(resp))
		w.Flush()
	}
	return nil
}
