package filter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/martian/v3"
	"shawnma.com/clarity/config"
	"shawnma.com/clarity/trie"
)

type Entry struct {
	Policy config.Policy
	// Temporary allowance
	ExpireTime     *time.Time
	UsedDuration   time.Duration
	LastAccessTime time.Time
}

type Filter struct {
	t *trie.PathTrie[*Entry]
	// skipped hosts
	skip *trie.PathTrie[bool]
}

func NewFilter(config *config.Config) *Filter {
	f := &Filter{}
	f.t = trie.NewPathTrie[*Entry]()

	for _, p := range config.Policies {
		log.Printf("Loading policy: %v", p)
		f.t.Put(p.Path, &Entry{Policy: p})
	}

	f.skip = trie.NewPathTrie[bool]()
	for _, h := range config.SkipProxy {
		h = strings.ReplaceAll(h, "*.", "")
		h = reverseHostname(h)
		f.skip.Put(h, true)
	}
	return f
}

func reverseHostname(h string) string {
	p := strings.Split(h, ".")
	ReverseSlice(p)
	return strings.Join(p, "/")
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
	h := reverseHostname(req.URL.Hostname())
	if f.skip.Search(h) {
		log.Printf("Skipping host %s", req.URL.Hostname())
		ctx.Session().SkipMitm()
		return nil
	}
	if req.Method == "CONNECT" || req.URL.Hostname() == "clarity.proxy" {
		return nil // proxy connect method, ignore.
	}
	url := req.URL
	path := url.Hostname() + url.Path
	// log.Printf("Filter: %s host %s", path, url.Hostname())
	err := f.t.WalkPath(path, func(key string, value *Entry) error {
		// log.Printf("walking %s", key)
		if value.ExpireTime != nil && value.ExpireTime.After(time.Now()) {
			// TODO: update last access time?
			log.Printf("path %s allowed as it is has not expired", key)
			return nil // we have a temp authorization
		}
		p := value.Policy
		for _, r := range p.AllowedRange {
			if r.InRange(time.Now()) {
				log.Printf("path %s allowed as it is in range: %v", key, r)
				return nil
			}
		}
		// rule matched, but neither is allowed, it must be denied
		return fmt.Errorf("rule denied at path %s when evaluating %+v", key, value)
	})
	if err != nil {
		log.Default().Println(err)
		conn, w, err := ctx.Session().Hijack()
		if err != nil {
			return err
		}
		resp := "HTTP/1.1 302 moved\nLocation: https://clarity.proxy/filter\nConnection: Close\n\n"
		w.Write([]byte(resp))
		w.Flush()
		conn.Close()
	}
	return nil
}

func ReverseSlice[T any](s []T) {
	first := 0
	last := len(s) - 1
	for first < last {
		s[first], s[last] = s[last], s[first]
		first++
		last--
	}
}
