package trie

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/transientvariable/hold"
	"github.com/transientvariable/hold/list"

	"github.com/stretchr/testify/assert"
)

func TestTrie_Add(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	t.Run("ab", func(t *testing.T) {
		value := "ab"
		err := trie.Add(value)
		assert.NoError(t, err)
		assert.Equal(t, trie.Len(), 1)
		assertContains(t, trie, value, true)
		assertContains(t, trie, "abc", false)
		assertContains(t, trie, "a", false)
		assertContains(t, trie, "acb", false)
	})

	t.Run("abcd", func(t *testing.T) {
		value := "abcd"
		err := trie.Add(value)
		assert.NoError(t, err)
		assert.Equal(t, trie.Len(), 2)
		assertContains(t, trie, value, true)
	})

	t.Run("acb", func(t *testing.T) {
		value := "acb"
		err := trie.Add(value)
		assert.NoError(t, err)
		assert.Equal(t, trie.Len(), 3)
		assertContains(t, trie, value, true)
	})

	t.Run("cbca", func(t *testing.T) {
		value := "cbca"
		err := trie.Add(value)
		assert.NoError(t, err)
		assert.Equal(t, trie.Len(), 4)
		assertContains(t, trie, value, true)
	})
}

func TestTrie_AddEntry(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	e := &entry{
		value: "dog",
		data:  "bark",
	}

	err = trie.AddEntry(e)
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 1)
	assertContains(t, trie, e.value, true)

	v, err := trie.Entry(e.value)
	assert.NoError(t, err)
	assert.Equal(t, e.value, v.Value())
	assert.Equal(t, e.data, v.Data())

	trie.Clear()
	assertSize(t, trie, 0)

	v, err = trie.Entry(e.value)
	assertError(t, err, hold.ErrCollectionEmpty)
	assert.Equal(t, nil, v)
}

func TestTrie_AddAll(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"the", "quick", "brown", "fox"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 4)
	assertContains(t, trie, "the", true)
	assertContains(t, trie, "quick", true)
	assertContains(t, trie, "brown", true)
	assertContains(t, trie, "fox", true)
	assertContentEquals(t, trie, "[brown, fox, quick, the]")
}

func TestTrie_Remove(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"jumped", "over", "the", "lazy", "dog"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 5)
	assertContains(t, trie, "jumped", true)
	assertContains(t, trie, "over", true)
	assertContains(t, trie, "the", true)
	assertContains(t, trie, "lazy", true)
	assertContains(t, trie, "dog", true)
	assertContentEquals(t, trie, "[dog, jumped, lazy, over, the]")

	r, err := trie.Remove("lazy")
	assert.NoError(t, err)
	assert.True(t, r, "expected result for removal of node 'lazy' to be true")

	r, err = trie.Remove("the")
	assert.NoError(t, err)
	assert.True(t, r, "expected result for removal of node 'the' to be true")

	r, err = trie.Remove("fox")
	assert.NoError(t, err)
	assert.False(t, r, "expected result for removal of node 'fox' to be false")
	assert.Equal(t, trie.Len(), 3)
	assertContains(t, trie, "lazy", false)
	assertContains(t, trie, "the", false)
	assertContentEquals(t, trie, "[dog, jumped, over]")

	trie.Clear()
	assertSize(t, trie, 0)
}

func TestTrie_MinMax(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"cba", "ab", "bce", "abcd"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 4)
	assertContentEquals(t, trie, "[ab, abcd, bce, cba]")

	tmin, err := trie.Min()
	assert.NoError(t, err)
	assertNodeValue(t, tmin, "ab")

	tmax, err := trie.Max()
	assert.NoError(t, err)
	assertNodeValue(t, tmax, "cba")
}

func TestTrie_Predecessor(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"bac", "dab", "dabb", "dac", "daca", "dabba", "ab"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 7)
	assertContentEquals(t, trie, "[ab, bac, dab, dabb, dabba, dac, daca]")

	p, err := trie.Predecessor("dabba")
	assert.NoError(t, err)
	assertNodeValue(t, p, "dabb")

	p, err = trie.Predecessor("bac")
	assert.NoError(t, err)
	assertNodeValue(t, p, "ab")
}

func TestTrie_Successor(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"bac", "dab", "dabb", "dac", "daca", "dabba", "ab"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 7)
	assertContentEquals(t, trie, "[ab, bac, dab, dabb, dabba, dac, daca]")

	s, err := trie.Successor("dabba")
	assert.NoError(t, err)
	assertNodeValue(t, s, "dac")

	s, err = trie.Successor("bac")
	assert.NoError(t, err)
	assertNodeValue(t, s, "dab")
}

func TestTrie_Completions(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"acb", "dabc", "daca", "da", "ab"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 5)
	assertContentEquals(t, trie, "[ab, acb, da, dabc, daca]")

	l := list.List[string]{}
	err = trie.Completions("a", &l)
	assert.NoError(t, err)
	assertContentEquals(t, &l, "[ab, acb]")

	l.Clear()
	err = trie.Completions("da", &l)
	assert.NoError(t, err)
	assertContentEquals(t, &l, "[da, dabc, daca]")
}

func TestTrie_LongestCommonPrefix(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"acb", "dadc", "dada", "da", "ab"})
	assert.NoError(t, err)
	assert.Equal(t, trie.Len(), 5)
	assertContentEquals(t, trie, "[ab, acb, da, dada, dadc]")

	l := list.List[string]{}
	err = trie.LongestCommonPrefix("a", &l)
	assert.NoError(t, err)
	assertContentEquals(t, &l, "[ab, acb]")

	l.Clear()
	err = trie.LongestCommonPrefix("dadda", &l)
	assert.NoError(t, err)
	assertContentEquals(t, &l, "[dada, dadc]")
}

func TestTrie_ValueAt(t *testing.T) {
	trie, err := New()
	assert.NoError(t, err)

	err = trie.AddAll(&list.List[string]{"Luffy", "Zoro", "Tony Chopper", "Sanji", "Frankie"})
	assert.NoError(t, err)
	assertContentEquals(t, trie, "[Frankie, Luffy, Sanji, Tony Chopper, Zoro]")

	entry, err := trie.ValueAt(2)
	assert.NoError(t, err)
	assert.Equal(t, "Sanji", entry.Value())
}

func assertError(t *testing.T, actual error, expected error) {
	t.Helper()

	if expected == nil && actual != nil {
		t.Errorf("expected error to be nil, but got '%s'", actual)
	} else {
		return
	}

	if errors.Is(actual, expected) {
		t.Errorf("expected error '%s', but got '%s'", expected, actual)
	}
}

func assertContains(t *testing.T, collection hold.Collection[string], value string, expected bool) {
	t.Helper()

	if collection.Contains(value) != expected {
		if expected {
			t.Errorf("expected to contain node: %s", value)
		} else {
			t.Errorf("expected not to contain node: %s", value)
		}
	}
}

func assertNodeValue(t *testing.T, actual any, expected string) {
	t.Helper()

	if n, ok := actual.(string); ok {
		if n != expected {
			t.Errorf("expected content of '%s', but found '%s'", expected, n)
		}
	} else {
		t.Errorf("expected type of 'string', but found '%v'", reflect.TypeOf(actual))
	}
}

func assertSize(t *testing.T, collection hold.Collection[string], expected int) {
	t.Helper()

	actual := collection.Len()
	if actual != expected {
		t.Errorf("expected size of '%d', but found '%d'", expected, actual)
	}
}

func assertContentEquals(t *testing.T, collection hold.Collection[string], expected string) {
	t.Helper()

	actual := fmt.Sprintf("%s", collection)
	if actual != expected {
		t.Errorf("expected content of '%s', but found '%s'", expected, actual)
	}
}
