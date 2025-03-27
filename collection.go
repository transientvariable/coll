package collection

// Iterator iterates over entries in a Collection.
type Iterator[E comparable] interface {
	// HasNext returns whether the iterator has more entries.
	HasNext() bool

	// Next returns the next entry in the iteration.
	//
	// If no further entries remain (HasNext() returns false), collection.ErrNoMoreElements should be returned.
	Next() (E, error)
}

// Collection defines the behavior for maintaining a collection of elements.
type Collection[E comparable] interface {
	// Add inserts the provided entries into the Collection.
	//
	// The returned error will be non-nil for bounded Collection implementations that have reached capacity and cannot
	// hold any further entries.
	Add(entry ...E) error

	// AddAll inserts all entries from the provided collection into the Collection.
	//
	// The returned error will be non-nil for bounded Collection implementations that have reached capacity and cannot
	// hold any further elements.
	AddAll(collection Collection[E]) error

	// Clear removes all elements from the Collection.
	Clear()

	// Contains returns true if an entry equivalent to the provided one exists in the Collection, otherwise false is
	// returned.
	Contains(entry E) bool

	// IsEmpty returns true if the Collection contains no elements, otherwise false is returned.
	IsEmpty() bool

	// Iterate returns the Iterator for the Collection.
	Iterate() Iterator[E]

	// Len returns the number of elements in the Collection.
	Len() int

	// Remove removes the first occurrence (if any) of an entry equivalent to the provided one.
	//
	// If an element was removed, the return value will be true, otherwise false will be returned.
	Remove(entry E) (bool, error)

	// Values returns a slice containing the elements in the Collection in the iteration order.
	Values() []E
}

// Ordered defines the behavior for a Collection whose elements are algorithmically positioned.
type Ordered[E comparable] interface {
	Collection[E]

	// Min returns the element with the lowest position in the Collection, which will be the first element in the
	// iteration order.
	Min() (E, error)

	// Max returns the element with the highest position in the Collection, which will be the last element in the
	// iteration order.
	Max() (E, error)

	// Predecessor returns the element before the first occurrence of the provided value in iteration order.
	Predecessor(entry E) (E, error)

	// Successor returns the element after the first occurrence of the provided element in iteration order.
	Successor(entry E) (E, error)
}

// Sequence defines the behavior for a container the represents a Collection of entries that are accessed via their
// position much like that of an array or slice.
type Sequence[E comparable] interface {
	Collection[E]

	// AddAt inserts the provided entry into the Sequence specified by index.
	//
	// The position of the entries that were at positions index to Sequence.Size() - 1 increase by one. The returned
	// error will be non-nil if the provided index is outside the current bounds of the Sequence
	// (index < 0 || index > Sequence.Size() - 1).
	AddAt(index int, entry E) error

	// AddFirst inserts the provided entry at the front (index == 0) of the Sequence.
	//
	// The positions of the existing entries are increased by one. The returned error will be non-nil for bounded
	// Sequence implementations that have reached capacity and cannot hold any further entries.
	AddFirst(entry E) error

	// AddLast inserts the provided entry at the end of the Sequence (index == Sequence.Size()).
	//
	// The returned error will be non-nil for bounded Sequence implementations that have reached capacity and cannot
	// hold any further entries.
	AddLast(entry E) error

	// Index returns the position of the first occurrence (if any) of an entry equivalent to the provided entry.
	//
	// The returned error will be non-nil if provided entry is not found in the Sequence, and the returned index will be
	// -1.
	Index(entry E) (int, error)

	// RemoveAt removes the entry at the provided index from the Sequence and returns it.
	//
	// The positions of the entries originally at positions index + 1 to Sequence.Size() - 1 are decremented by 1. The
	// returned error will be non-nil if the provided index is outside the bounds of the Sequence
	// (index < 0 || index > Sequence.Size() - 1).
	RemoveAt(index int) (E, error)

	// RemoveFirst removes the entry at the front (index == 0) of the Sequence and returns it.
	//
	// If the Sequence is empty (Sequence.Size() == 0), the return value will be nil.
	RemoveFirst() (E, error)

	// RemoveLast removes the entry at the end (index == Sequence.Size() - 1) of the Sequence and returns it.
	//
	// If the Sequence is empty (Sequence.Size() == 0), the return value will be nil.
	RemoveLast() (E, error)

	// ValueAt returns the entry at the position specified by the provided index.
	//
	// The returned error will be non-nil if the provided index is outside the current bounds of the Sequence
	// (index < 0 || index > Sequence.Size() - 1).
	ValueAt(index int) (E, error)
}
