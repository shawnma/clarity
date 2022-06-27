package filter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/martian/v3"
	"gopkg.in/yaml.v3"
	"shawnma.com/clarity/trie"
)

// Overall policy for a path
type Policy struct {
	Path string

	// If configured, only the allowed time range will be permitted to access this website
	// Otherwise, it will be always allowed unless reaches the MaxAllowed duration
	AllowedRange []TimeRange

	// Max duration allowed for this website.
	MaxAllowed time.Duration

	// If true, the website will be self-managed up to MaxAllowed duration;
	// otherwise, the MaxAllowed will be ignored and the website is allowed during the TimeRanges
	SelfManaged bool
}

type Entry struct {
	Policy Policy
	// Temporary allowance
	ExpireTime     *time.Time
	UsedDuration   time.Duration
	LastAccessTime time.Time
}

type Filter struct {
	t *trie.PathTrie[*Entry]
}

func NewFilter() *Filter {
	f := &Filter{}
	f.t = trie.NewPathTrie[*Entry]()
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Unable to open config file config.yaml")
	}
	var polices []Policy
	err = yaml.Unmarshal(data, &polices)
	if err != nil {
		log.Fatalf("Unable to parse config: %s", err)
	}
	for _, p := range polices {
		log.Printf("Loading policy: %v", p)
		f.t.Put(p.Path, &Entry{Policy: p})
	}
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
	if req.Method == "CONNECT" {
		return nil // proxy connect method, ignore.
	}
	ctx := martian.NewContext(req)
	url := req.URL
	path := url.Hostname() + url.Path
	log.Printf("Filter: %s", path)
	err := f.t.WalkPath(path, func(key string, value *Entry) error {
		log.Printf("walking %s", key)
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
