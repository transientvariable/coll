package list

import (
	"errors"
	"fmt"
	"testing"

	"github.com/transientvariable/collection"

	"github.com/stretchr/testify/assert"
)

type entry struct {
	value    any
	position int
}

func TestAdd(t *testing.T) {
	entries := []entry{
		{value: "piranha plant", position: 0},
		{value: "samus", position: 1},
		{value: "jigglypuff", position: 2},
		{value: "r.o.b.", position: 3},
		{value: "mega man", position: 4},
		{value: "yoshi", position: 5},
	}

	t.Run("Add", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		entry := entry{value: "gumball", position: list.Len()}
		err := list.Add(entry)

		assertError(t, err, nil)
		assertSize(t, list, 7)
		assertContains(t, &list, entry, true)
		assertIndex(t, list, entry, 6)
	})

	t.Run("AddFirst", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		entry := entry{value: "luffy", position: 0}
		err := list.AddFirst(entry)

		assertError(t, err, nil)
		assertSize(t, list, 7)
		assertContains(t, &list, entry, true)
		assertIndex(t, list, entry, 0)
	})

	t.Run("AddLast", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		entry := entry{value: "snorlax", position: list.Len()}
		err := list.AddLast(entry)

		assertError(t, err, nil)
		assertSize(t, list, 7)
		assertContains(t, &list, entry, true)
		assertIndex(t, list, entry, 0)
	})

	t.Run("AddAt", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		index := 3
		entry := entry{value: "chopper", position: index}
		err := list.AddAt(index, entry)

		assertError(t, err, nil)
		assertSize(t, list, 7)
		assertContains(t, &list, entry, true)
		assertIndex(t, list, entry, index)
	})

	t.Run("AddAll", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)

		newElements := []entry{
			{value: "gumball", position: 6},
			{value: "luffy", position: 7},
			{value: "chopper", position: 8},
		}

		newList := List[entry]{}
		newList = append(newList, newElements...)
		err := list.AddAll(&newList)

		assertError(t, err, nil)
		assertSize(t, list, 9)

		assertContains(t, &list, newElements[0], true)
		assertIndex(t, list, newElements[0], newElements[0].position)

		assertContains(t, &list, newElements[1], true)
		assertIndex(t, list, newElements[1], newElements[1].position)

		assertContains(t, &list, newElements[2], true)
		assertIndex(t, list, newElements[2], newElements[2].position)
	})
}

func TestRemove(t *testing.T) {
	entries := []entry{
		{value: "piranha plant", position: 0},
		{value: "samus", position: 1},
		{value: "jigglypuff", position: 2},
		{value: "r.o.b.", position: 3},
		{value: "mega man", position: 4},
		{value: "yoshi", position: 5},
	}

	t.Run("Remove", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		entry := entries[3]

		r, err := list.Remove(entry)
		assertError(t, err, nil)

		if !r {
			t.Error("expected result to be true")
		}

		r, err = list.Remove(entry)
		assertError(t, err, nil)

		if r {
			t.Error("expected result to be false")
		}

		assertSize(t, list, 5)
		assertContains(t, &list, entry, false)
	})

	t.Run("RemoveFirst", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)

		entry, err := list.RemoveFirst()
		assertError(t, err, nil)

		assert.Equal(t,
			entries[0].value,
			entry.value,
			fmt.Sprintf("expected value '%s' for first entry, but found '%s'", entries[0].value, entry.value),
		)

		assertSize(t, list, 5)
		assertContains(t, &list, entry, false)
		assertIndex(t, list, entries[1], 0)
	})

	t.Run("RemoveLast", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)

		entry, err := list.RemoveLast()
		assertError(t, err, nil)

		assert.Equal(t,
			entries[len(entries)-1].value,
			entry.value,
			fmt.Sprintf("expected value '%s' for last entry, but found '%s'", entries[0].value, entry.value),
		)

		assertSize(t, list, 5)
		assertContains(t, &list, entry, false)
	})

	t.Run("RemoveAt", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)

		entry, err := list.RemoveAt(4)
		assertError(t, err, nil)

		assertSize(t, list, 5)
		assertContains(t, &list, entry, false)
		assertIndex(t, list, entries[5], 4)
	})

	t.Run("Clear", func(t *testing.T) {
		list := List[entry]{}
		list = append(list, entries...)
		assertSize(t, list, 6)

		if list.IsEmpty() {
			t.Error("expected result to be false")
		}

		list.Clear()

		if !list.IsEmpty() {
			t.Error("expected result to be true")
		}
	})
}

func assertContains(t *testing.T, collection collection.Collection[entry], value entry, expected bool) {
	t.Helper()
	if collection.Contains(value) != expected {
		if expected {
			t.Errorf("expected to contain value: %+v", value)
		} else {
			t.Errorf("expected not to contain value: %+v", value)
		}
	}
}
func assertError(t *testing.T, actual error, expected error) {
	t.Helper()
	if !errors.Is(actual, expected) {
		t.Errorf("expected error '%s', but found '%s'", expected, actual)
	}
}
func assertIndex(t *testing.T, list List[entry], value entry, expected int) {
	t.Helper()
	actual, err := list.Index(value)
	if err != nil {
		t.Errorf("expected index of '%d', but found '%d'", expected, actual)
	}
}
func assertSize(t *testing.T, list List[entry], expected int) {
	t.Helper()
	actual := list.Len()
	if actual != expected {
		t.Errorf("expected size of '%d', but found '%d'", expected, actual)
	}
}
