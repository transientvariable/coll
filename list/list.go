package list

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/transientvariable/collection-go"
)

var _ collection.Sequence[any] = (*List[any])(nil)

type iterator[E comparable] struct {
	index int
	list  List[E]
}

func (i *iterator[E]) HasNext() bool {
	if i.list.Len() == 0 {
		return false
	}
	return i.index < i.list.Len()
}

func (i *iterator[E]) Next() (E, error) {
	var n E
	if !i.HasNext() {
		return n, fmt.Errorf("list_iter: %w", collection.ErrNoMoreElements)
	}
	n, err := i.list.ValueAt(i.index)
	if err != nil {
		return n, err
	}
	i.index++
	return n, nil
}

// List is a basic implementation of a Sequence.
//
// This implementation does not make any guarantees for concurrent access.
type List[E comparable] []E

// Add inserts the provided entry into the List.
func (l *List[E]) Add(entry ...E) error {
	*l = append(*l, entry...)
	return nil
}

// AddAll inserts all entries from the provided List into the List.
func (l *List[E]) AddAll(collection collection.Collection[E]) error {
	if collection != nil {
		*l = append(*l, collection.Values()...)
	}
	return nil
}

// AddAt inserts the provided entry into the List specified by index.
//
// The position of the entries that were at positions index to List.Size() - 1 increase by one. The returned error will
// be non-nil if the provided index is outside the current bounds of the List (index < 0 || index > List.Size() - 1).
func (l *List[E]) AddAt(index int, entry E) error {
	if err := l.checkBounds(index); err != nil {
		return err
	}
	var e E
	*l = append(*l, e)
	copy((*l)[index+1:], (*l)[index:])
	(*l)[index] = entry
	return nil
}

// AddFirst inserts the provided value at the front (index == 0) of the List.
//
// The positions of the existing entries are increased by one.
func (l *List[E]) AddFirst(value E) error {
	*l = append([]E{value}, *l...)
	return nil
}

// AddLast inserts the provided value at the end of the List (index == List.Size()).
func (l *List[E]) AddLast(value E) error {
	return l.Add(value)
}

// Clear removes all entries from the List.
func (l *List[E]) Clear() {
	*l = List[E]{}
}

// Contains returns true if an entry equivalent to the provided value exists in the List, otherwise false is
// returned.
func (l *List[E]) Contains(value E) bool {
	if _, err := l.Index(value); err == nil {
		return true
	}
	return false
}

// Index returns the position of the first occurrence (if any) of an entry equivalent to the provided entry.
//
// The returned error will be non-nil if provided entry is not found in the List, and the returned index will be equal
// to collection.ErrNotFound.
func (l *List[E]) Index(value E) (int, error) {
	i, err := l.findFirst(value)
	if err != nil {
		return i, err
	}
	return i, nil
}

// IsEmpty returns true if the List contains no entries, otherwise false is returned.
func (l *List[E]) IsEmpty() bool {
	return l.Len() == 0
}

// Iterate returns the collection.Iterator for the List.
func (l *List[E]) Iterate() collection.Iterator[E] {
	return &iterator[E]{list: *l}
}

// Len returns the number of entries in the List.
func (l *List[E]) Len() int {
	return len(*l)
}

// Remove removes the first occurrence (if any) of an entry equivalent to the provided value.
//
// If an entry was removed, the return value will be true, otherwise false will be returned.
func (l *List[E]) Remove(value E) (bool, error) {
	if l.Contains(value) {
		i, _ := l.Index(value)
		_, err := l.RemoveAt(i)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// RemoveAt removes the entry at the provided index from the List and returns it.
//
// The positions of the entries original positions index + 1 to List.Size() - 1 are decremented by 1. The returned error
// will be non-nil if the provided index is outside the bounds of the List (index < 0 || index > List.Size() - 1).
func (l *List[E]) RemoveAt(index int) (E, error) {
	entry, err := l.ValueAt(index)
	if err != nil {
		return entry, err
	}
	copy((*l)[index:l.Len()-1], (*l)[index+1:l.Len()])
	*l = (*l)[:l.Len()-1]
	return entry, nil
}

// RemoveFirst removes the entry at the front (index == 0) of the List and returns it.
//
// If the List is empty (List.Size() == 0), the return value will be nil.
func (l *List[E]) RemoveFirst() (E, error) {
	var entry E
	if l.Len() > 0 {
		e, err := l.RemoveAt(0)
		if err != nil {
			return entry, err
		}
		entry = e
	}
	return entry, nil
}

// RemoveLast removes the entry at the end (index == List.Size() - 1) of the List and returns it.
//
// If the List is empty (List.Size() == 0), the return value will be nil.
func (l *List[E]) RemoveLast() (E, error) {
	var entry E
	if l.Len() > 0 {
		e, err := l.RemoveAt(l.Len() - 1)
		if err != nil {
			return entry, err
		}
		entry = e
	}
	return entry, nil
}

// ValueAt returns the entry at the position specified by the provided index.
//
// The returned error will be non-nil if the provided index is outside the current bounds of the List
// (index < 0 || index > List.Size() - 1).
func (l *List[E]) ValueAt(index int) (E, error) {
	if err := l.checkBounds(index); err != nil {
		var e E
		return e, err
	}
	return (*l)[index], nil
}

// Values returns a slice containing the entries in the List in the iteration order.
func (l *List[E]) Values() []E {
	entries := make([]E, l.Len())
	copy(entries, *l)
	return entries
}

// String returns a string representation of the List in it's current state.
func (l *List[E]) String() string {
	if l.Len() == 0 {
		return "[]"
	}

	entries := make([]string, 0, l.Len())
	for _, e := range *l {
		entries = append(entries, fmt.Sprintf("%v", e))
	}
	return "[" + strings.Join(entries, ", ") + "]"
}

func (l *List[E]) checkBounds(index int) error {
	if index < 0 || index > l.Len() {
		return fmt.Errorf("list: size = %d, requested index = %d: %w", l.Len(), index, collection.ErrBoundsOutOfRange)
	}
	return nil
}

func (l *List[E]) findFirst(entry E) (int, error) {
	for i, v := range *l {
		if reflect.DeepEqual(v, entry) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("list: %w", collection.ErrNotFound)
}
