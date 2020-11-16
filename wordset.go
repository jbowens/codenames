package codenames

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

type wordSetID [sha1.Size]byte

func (i wordSetID) String() string {
	return fmt.Sprintf("%x", i[:])
}

type WordSets struct {
	mu   sync.Mutex
	byID map[wordSetID][]string
}

func (ws *WordSets) init() {
	if ws.byID == nil {
		ws.byID = make(map[wordSetID][]string)
	}
}

func (ws *WordSets) Canonicalize(words []string) (wordSetID, []string, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.init()

	set := map[string]bool{}
	for _, w := range words {
		set[strings.TrimSpace(strings.ToUpper(w))] = true
	}
	if len(set) > 0 && len(set) < 25 {
		return wordSetID{}, nil, errors.New("need at least 25 words")
	}

	words = words[:0]
	for w := range set {
		words = append(words, w)
	}
	sort.Strings(words)

	// Calculate the word set ID, a hash of the canonicalized word set.
	h := sha1.New()
	for _, w := range words {
		io.WriteString(h, w)
		h.Write([]byte{0x00})
	}
	idBytes := h.Sum(nil)
	var id wordSetID
	copy(id[:], idBytes)

	if interned, ok := ws.byID[id]; ok {
		return id, interned, nil
	}
	ws.byID[id] = words
	return id, words, nil
}
