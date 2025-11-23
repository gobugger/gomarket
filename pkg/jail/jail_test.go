package jail

import (
	"fmt"
	"testing"
)

func TestMakeIndexes(t *testing.T) {
	text := "a9bpojrj2ro2cllkeej2x00ts2an1dxjesoveemudpdygaaddxiiuuzd.onion"
	indexes := makeIndexes(text, 4)

	fmt.Println(indexes)

	if len(indexes) != 4 {
		t.Fatal("wrong number of indexes")
	}

	m := map[int]bool{}
	prev := indexes[0]
	for _, idx := range indexes {
		if idx >= len(text) {
			t.Fatal("invalid index")
		}
		m[idx] = true
		if idx < prev {
			t.Fatalf("indexes should be sorted %d %d", prev, idx)
		}
		prev = idx
	}

	if len(indexes) != len(m) {
		t.Fatal("there should be no duplicate indexes")
	}
}
