// Refactored from https://github.com/dghubble/trie/blob/master/path_trie.go

package trie

import (
	"strings"
)

// PathTrie is a trie of paths with string keys and interface{} values.

// PathTrie is a trie of string keys and interface{} values. Internal nodes
// have nil values so stored nil values cannot be distinguished and are
// excluded from walks. By default, PathTrie will segment keys by forward
// slashes with PathSegmenter (e.g. "/a/b/c" -> "/a", "/b", "/c"). A custom
// StringSegmenter may be used to customize how strings are segmented into
// nodes. A classic trie might segment keys by rune (i.e. unicode points).
type PathTrie[T comparable] struct {
	value    T
	children map[string]*PathTrie[T]
}

// WalkFunc defines some action to take on the given key and value during
// a Trie Walk. Returning a non-nil error will terminate the Walk.
type WalkFunc[T any] func(key string, value T) error

// PathSegmenter segments string key paths by slash separators. For example,
// "/a/b/c" -> ("/a", 2), ("/b", 4), ("/c", -1) in successive calls. It does
// not allocate any heap memory.
func pathSegmenter(path string, start int) (segment string, next int) {
	if len(path) == 0 || start < 0 || start > len(path)-1 {
		return "", -1
	}
	end := strings.IndexRune(path[start+1:], '/') // next '/' after 0th rune
	if end == -1 {
		return path[start:], -1
	}
	return path[start : start+end+1], start + end + 1
}

// NewPathTrie allocates and returns a new *PathTrie.
func NewPathTrie[T comparable]() *PathTrie[T] {
	return &PathTrie[T]{}
}

// Get returns the value stored at the given key. Returns nil for internal
// nodes or for nodes with a value of nil.
func (trie *PathTrie[T]) Get(key string) (t T) {
	node := trie
	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
		node = node.children[part]
		if node == nil {
			return
		}
	}
	return node.value
}

// Search returns the value stored at the given key. Returns nil for internal
// nodes or for nodes with a value of nil.
func (trie *PathTrie[T]) Search(key string) (t T) {
	node := trie
	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
		child := node.children[part]
		if child == nil {
			return node.value
		}
		node = child
	}
	return node.value
}

// Put inserts the value into the trie at the given key, replacing any
// existing items. It returns true if the put adds a new value, false
// if it replaces an existing value.
// Note that internal nodes have nil values so a stored nil value will not
// be distinguishable and will not be included in Walks.
func (trie *PathTrie[T]) Put(key string, value T) bool {
	node := trie
	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
		child := node.children[part]
		if child == nil {
			if node.children == nil {
				node.children = map[string]*PathTrie[T]{}
			}
			child = NewPathTrie[T]()
			node.children[part] = child
		}
		node = child
	}
	// does node have an existing value?
	isNewVal := node.value == *new(T)
	node.value = value
	return isNewVal
}

// Delete removes the value associated with the given key. Returns true if a
// node was found for the given key. If the node or any of its ancestors
// becomes childless as a result, it is removed from the trie.
func (trie *PathTrie[T]) Delete(key string) bool {
	var path []nodeStr[T] // record ancestors to check later
	node := trie
	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
		path = append(path, nodeStr[T]{part: part, node: node})
		node = node.children[part]
		if node == nil {
			// node does not exist
			return false
		}
	}
	// delete the node value
	node.value = *new(T)
	// if leaf, remove it from its parent's children map. Repeat for ancestor path.
	if node.isLeaf() {
		// iterate backwards over path
		for i := len(path) - 1; i >= 0; i-- {
			parent := path[i].node
			part := path[i].part
			delete(parent.children, part)
			if !parent.isLeaf() {
				// parent has other children, stop
				break
			}
			parent.children = nil
			if parent.value != *new(T) {
				// parent has a value, stop
				break
			}
		}
	}
	return true // node (internal or not) existed and its value was nil'd
}

// Walk iterates over each key/value stored in the trie and calls the given
// walker function with the key and value. If the walker function returns
// an error, the walk is aborted.
// The traversal is depth first with no guaranteed order.
func (trie *PathTrie[T]) Walk(walker WalkFunc[T]) error {
	return trie.walk("", walker)
}

// WalkPath iterates over each key/value in the path in trie from the root to
// the node at the given key, calling the given walker function for each
// key/value. If the walker function returns an error, the walk is aborted.
func (trie *PathTrie[T]) WalkPath(key string, walker WalkFunc[T]) error {
	// Get root value if one exists.
	if trie.value != *new(T) {
		if err := walker("", trie.value); err != nil {
			return err
		}
	}
	for part, i := pathSegmenter(key, 0); ; part, i = pathSegmenter(key, i) {
		if trie = trie.children[part]; trie == nil {
			return nil
		}
		if trie.value != *new(T) {
			var k string
			if i == -1 {
				k = key
			} else {
				k = key[0:i]
			}
			if err := walker(k, trie.value); err != nil {
				return err
			}
		}
		if i == -1 {
			break
		}
	}
	return nil
}

// PathTrie node and the part string key of the child the path descends into.
type nodeStr[T comparable] struct {
	node *PathTrie[T]
	part string
}

func (trie *PathTrie[T]) walk(key string, walker WalkFunc[T]) error {
	if trie.value != *new(T) {
		if err := walker(key, trie.value); err != nil {
			return err
		}
	}
	for part, child := range trie.children {
		if err := child.walk(key+part, walker); err != nil {
			return err
		}
	}
	return nil
}

func (trie *PathTrie[T]) isLeaf() bool {
	return len(trie.children) == 0
}
