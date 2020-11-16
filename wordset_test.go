package codenames

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestWordSetCanonicalize(t *testing.T) {
	b, err := ioutil.ReadFile("frontend/words.json")
	if err != nil {
		t.Fatal(err)
	}
	var defaultWordsets map[string][]string
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&defaultWordsets)
	if err != nil {
		t.Fatal(err)
	}

	internedSets := map[string][]string{}

	var ws WordSets
	for name, words := range defaultWordsets {
		id, interned, err := ws.Canonicalize(words)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s : %s\n", name, id)
		internedSets[name] = interned
	}

	for name, words := range defaultWordsets {
		words2 := append([]string{}, words...)
		_, interned, err := ws.Canonicalize(words2)
		if err != nil {
			t.Fatal(err)
		}
		if &internedSets[name][0] != &interned[0] {
			t.Errorf("word set %q has different slice pointer 2nd canonicalization", name)
		}
	}
}
