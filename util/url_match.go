package util

import (
	"log"
	"strings"
)

// UrlMatch matches a url using the "suffix" of its domain and "prefix" of its
// path. It will return the matched value.
// For example, www.client6.google.com/chat/log will match google.com/chat
type UrlMatch[T comparable] struct {
	paths PathTrie[bool]
	t     PathTrie[T]
}

func (u *UrlMatch[T]) Add(url string, t T) {
	h, p := splitUrl(url)
	h = reverseHost(h)
	log.Printf("Adding %s %s", h, p)
	u.paths.Put(h, true)
	u.t.Put(h+p, t)
}

func (u *UrlMatch[T]) Match(host, path string) (t T) {
	host = reverseHost(host)
	def := *new(T)
	_ = u.paths.WalkPath(host, func(key string, value bool) error {
		if value {
			v := u.t.Search(key + path)
			if v != def {
				t = v
			}
		}
		return nil
	})
	return t
}

func reverseHost(host string) string {
	parts := strings.Split(host, ".")
	parts = ReverseSlice(parts)
	return strings.Join(parts, "/")
}

func splitUrl(url string) (string, string) {
	h := strings.Index(url, "/")
	if h > 0 {
		return url[:h], url[h:]
	} else {
		return url, ""
	}
}
