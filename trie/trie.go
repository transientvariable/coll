package trie

import (
	"fmt"
	"io"
	"strings"

	"github.com/transientvariable/anchor"
	"github.com/transientvariable/hold"
	"github.com/transientvariable/hold/list"
)

var _ Trie = (*trie)(nil)

// Entry is a container for entries that can be inserted into a Trie.
type Entry interface {
	Value() string
	Data() any
}

// NewEntry creates a generic Entry that can be used with a Trie.
func NewEntry(value string, data any) Entry {
	return &entry{value: value, data: data}
}

type entry struct {
	value string
	data  any
}

func (e *entry) Value() string {
	return e.value
}

func (e *entry) Data() any {
	return e.data
}

func (e *entry) String() string {
	return e.Value()
}

// Trie ...
type Trie interface {
	hold.Ordered[string]

	// AddEntry inserts the provided Entry into the Trie.
	//
	// The returned error will be non-nil if the Trie has reached capacity and cannot hold any further entries.
	AddEntry(entry Entry) error

	// AddAllEntries inserts the provided collection of entries into the Trie.
	//
	// The returned error will be non-nil if the Trie has reached capacity and cannot hold any further entries.
	AddAllEntries(entries hold.Collection[Entry]) error

	// Completions finds all entries in the Trie that match the provided prefix, and appends the matching entries
	// (if any) to the provided collection.
	Completions(prefix string, entries hold.Collection[string]) error

	// Entry returns the entry corresponding to the provided value.
	//
	// The returned error will be non-nil if:
	//   - the Trie is empty (has no elements)
	//   - the value provided for locating an Entry is blank
	//   - the Trie does not contain an Entry corresponding to the provided value
	Entry(value string) (Entry, error)

	// Entries returns a slice containing the entries in the Trie in iteration order.
	Entries() ([]Entry, error)

	// Leaves returns all the entries that are immediate children of the Entry matching the provided value.
	//
	// The returned error will be non-nil if:
	//   - the Trie is empty (has no elements)
	//   - the value provided for locating an Entry is blank
	//   - the Trie does not contain an Entry corresponding to the provided value
	Leaves(value string) ([]Entry, error)

	// LongestCommonPrefix finds all entries in the Trie that share the longest common prefix with the provided prefix,
	// and appends the matching entries (if any) to the provided collection.
	LongestCommonPrefix(prefix string, entries hold.Collection[string]) error

	// RemoveEntry removes the first occurrence (if any) of an entry corresponding to the provided Entry.
	//
	// If an entry was removed, the return node will be true, otherwise false will be returned.
	RemoveEntry(entry Entry) (bool, error)

	// ValueAt returns the entry at the position specified by the provided index.
	//
	// The returned error will be non-nil if the provided index is outside the current bounds of the Trie
	// (index < 0 || index > trie.Size() - 1).
	ValueAt(index int) (Entry, error)
}

type trie struct {
	digitizer Digitizer
	head      Leaf
	root      Node
	size      int
	tail      Leaf
}

// New creates a new Trie with the provided options.
func New(options ...func(*Option)) (Trie, error) {
	opts := &Option{}
	for _, opt := range options {
		opt(opts)
	}

	head := &leaf{
		node:   newNode(0),
		isHead: true,
	}

	tail := &leaf{
		node:   newNode(0),
		isTail: true,
	}

	head.SetNext(tail)
	tail.SetNext(head)

	trie := &trie{
		digitizer: NewASCIIDigitizer(),
		head:      head,
		tail:      tail,
	}

	if opts.digitizer != nil {
		if opts.digitizer.Base() <= 0 {
			return nil, fmt.Errorf("trie: base for digitizer must be greater than 0")
		}
		trie.digitizer = opts.digitizer
	}
	return trie, nil
}

// Add inserts the provided node into the Trie. The returned error will be non-nil if the Trie has reached capacity and
// cannot hold any further entries.
func (t *trie) Add(values ...string) error {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			if err := t.AddEntry(&entry{value: v}); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddAll inserts all values from the provided collection into the Trie. The returned error will be non-nil if the Trie
// has reached capacity and cannot hold any further entries.
func (t *trie) AddAll(values hold.Collection[string]) error {
	entries := list.List[Entry]{}
	if values != nil {
		for _, v := range values.Values() {
			if v = strings.TrimSpace(v); v == "" {
				continue
			}

			if err := entries.Add(&entry{value: v}); err != nil {
				return err
			}
		}
	}
	return t.AddAllEntries(&entries)
}

// AddEntry inserts the provided Entry into the Trie.
//
// The returned error will be non-nil if the Trie has reached capacity and cannot hold any further entries.
func (t *trie) AddEntry(entry Entry) error {
	_, err := t.insert(entry)
	return err
}

// AddAllEntries inserts the provided collection of entries into the Trie. The returned error will be non-nil if the
// Trie has reached capacity and cannot hold any further entries.
func (t *trie) AddAllEntries(entries hold.Collection[Entry]) error {
	if entries != nil {
		for _, v := range entries.Values() {
			if err := t.AddEntry(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Clear removes all entries from the Trie.
func (t *trie) Clear() {
	iter := newIterator(t, t.head)
	for iter.advance() {
		_ = iter.remove()
	}
}

// Completions finds all entries in the Trie that match the provided prefix, and appends the matching entries (if any)
// to the provided collection.
func (t *trie) Completions(prefix string, entries hold.Collection[string]) error {
	if t.IsEmpty() {
		return fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	searchResult, err := t.find(ctx, prefix)
	if err != nil {
		return err
	}

	numDigits := t.digitizer.NumDigitsOf(prefix)
	if t.digitizer.IsPrefixFree() {
		numDigits--
		eos, err := ctx.processedEndOfString(prefix)
		if err != nil {
			return err
		}

		if eos {
			ctx.ascend()
		}
	}

	if searchResult == Prefix || searchResult == Matched || ctx.branchPosition == numDigits {
		if err := ctx.entriesInSubtree(entries); err != nil {
			return err
		}
	}
	return nil
}

// Contains returns true if an entry equivalent to the provided node exists in the Trie, otherwise false is returned.
func (t *trie) Contains(value string) bool {
	if t.IsEmpty() {
		return false
	}

	if value = strings.TrimSpace(value); value == "" {
		return false
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	r, err := t.find(ctx, value)
	if err != nil {
		return false
	}
	return r == Matched
}

// Entries returns a slice containing the entries in the Trie in iteration order.
func (t *trie) Entries() ([]Entry, error) {
	var entries []Entry
	iter := newIterator(t, t.head)
	for iter.advance() {
		entry, err := iter.get()
		if err != nil {
			return entries, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Entry returns the entry corresponding to the provided node. The returned error will be non-nil if:
//   - the Trie is empty (has no elements)
//   - the node provided for locating an Entry is blank
//   - the Trie does not contain an Entry corresponding to the provided node
func (t *trie) Entry(value string) (Entry, error) {
	v, err := t.node(value)
	if err != nil {
		return nil, err
	}
	return v.Value(), nil
}

// IsEmpty returns true if the Trie contains no entries, otherwise false is returned.
func (t *trie) IsEmpty() bool {
	return t.Len() == 0
}

// Iterate returns the collection.Iterator for the Trie.
func (t *trie) Iterate() hold.Iterator[string] {
	return newIterator(t, t.head)
}

// Leaves returns all the entries that are immediate children of the Entry matching the provided value. The returned
// error will be non-nil if:
//   - the Trie is empty (has no elements)
//   - the value provided for locating an Entry is blank
//   - the Trie does not contain an Entry corresponding to the provided value
func (t *trie) Leaves(value string) ([]Entry, error) {
	n, err := t.node(value)
	if err != nil {
		return nil, err
	}

	var leaves []Entry
	for _, c := range n.Children() {
		if c.IsLeaf() && c.Value() != nil {
			leaves = append(leaves, c.Value())
		}
	}
	return leaves, nil
}

// Len returns the number of entries in the Trie.
func (t *trie) Len() int {
	return t.size
}

// LongestCommonPrefix finds all entries in the Trie that share the longest common prefix with the provided prefix,
// and appends the matching entries (if any) to the provided collection.
func (t *trie) LongestCommonPrefix(prefix string, entries hold.Collection[string]) error {
	if t.IsEmpty() {
		return fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	_, err := t.find(ctx, prefix)
	if err != nil {
		return err
	}

	eos, err := ctx.processedEndOfString(prefix)
	if err != nil {
		return err
	}

	if eos {
		ctx.ascend()
	}

	if err := ctx.entriesInSubtree(entries); err != nil {
		return err
	}
	return nil
}

// Min returns the entry with the lowest position in the Trie. More specifically, the first entry in the iteration
// order is returned.
func (t *trie) Min() (string, error) {
	if t.IsEmpty() {
		return "", fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}
	return t.head.Next().Value().Value(), nil
}

// Max returns the entry with the highest position in the Trie. More specifically, the last entry in the iteration
// order is returned.
func (t *trie) Max() (string, error) {
	if t.IsEmpty() {
		return "", fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}
	return t.tail.Previous().Value().Value(), nil
}

// Predecessor returns the entry (if any) from the Trie that is less than the provided node. More specifically, the
// entry before the first occurrence of the provided entry in iteration order is returned.
func (t *trie) Predecessor(value string) (string, error) {
	if t.IsEmpty() {
		return value, fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	if value = strings.TrimSpace(value); value == "" {
		return value, fmt.Errorf("trie: %w", hold.ErrValueRequired)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	r, err := t.find(ctx, value)
	if err != nil {
		return value, err
	}

	m, err := t.moveToPredecessor(ctx, value, r)
	if err != nil {
		return "", err
	}

	if m {
		return ctx.pointer.Value().Value(), nil
	}
	return value, fmt.Errorf("trie: %w", hold.ErrNotFound)
}

// Remove removes the first occurrence (if any) of an entry equivalent to the provided node. If an entry was
// removed, the return node will be true, otherwise false will be returned.
func (t *trie) Remove(value string) (bool, error) {
	return t.RemoveEntry(&entry{value: value})
}

// RemoveEntry removes the first occurrence (if any) of an entry corresponding to the provided Entry. If an entry
// was removed, the return node will be true, otherwise false will be returned.
func (t *trie) RemoveEntry(entry Entry) (bool, error) {
	if t.IsEmpty() {
		return false, fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	if entry == nil {
		return false, fmt.Errorf("trie: %w", hold.ErrValueRequired)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	r, err := t.find(ctx, entry.Value())
	if err != nil {
		return false, nil
	}

	if r != Matched {
		return false, nil
	}

	if err := t.remove(ctx.pointer); err != nil {
		return false, err
	}
	return true, nil
}

// Successor returns the entry (if any) from the Trie that is greater than the provided node. More specifically, the
// entry after the first occurrence of the provided node in iteration order is returned.
func (t *trie) Successor(value string) (string, error) {
	if t.IsEmpty() {
		return value, fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	if value = strings.TrimSpace(value); value == "" {
		return value, fmt.Errorf("trie: %w", hold.ErrValueRequired)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	r, err := t.find(ctx, value)
	if err != nil {
		return value, err
	}

	successor := t.tail
	if r == Matched {
		successor = ctx.pointer.(Leaf).Next()
	} else {
		r, err := t.find(ctx, value)
		if err != nil {
			return value, err
		}

		m, err := t.moveToPredecessor(ctx, value, r)
		if err != nil {
			return "", err
		}

		if m {
			successor = ctx.pointer.(Leaf).Next()
		}
	}

	if !successor.IsTail() {
		return successor.Value().Value(), nil
	}
	return value, fmt.Errorf("trie: %w", hold.ErrNotFound)
}

// ValueAt returns the entry at the position specified by the provided index. The returned error will be
// non-nil if the provided index is outside the current bounds of the trie (index < 0 || index > trie.Size() - 1).
func (t *trie) ValueAt(index int) (Entry, error) {
	if err := t.checkBounds(index); err != nil {
		return nil, err
	}

	iter := newIterator(t, t.head)

	var i int
	for iter.HasNext() {
		v, err := iter.Next()
		if err != nil {
			return nil, err
		}

		if i == index {
			return t.Entry(v)
		}
		i++
	}
	return nil, io.EOF
}

// Values returns a slice containing the values for each Entry in the Trie in iteration order.
func (t *trie) Values() []string {
	entries, err := t.Entries()
	if err != nil {
		panic(err)
	}

	values := make([]string, len(entries))
	for i, e := range entries {
		values[i] = e.Value()
	}
	return values
}

// String returns a string representation of the Trie in its current state.
func (t *trie) String() string {
	if t.Len() == 0 {
		return "{}"
	}

	m := make(map[string]any)
	iter := newIterator(t, t.head)
	for iter.advance() {
		entry, err := iter.get()
		if err != nil {
			return fmt.Sprintf(`{error: "%s"}`, err.Error())
		}
		m[entry.Value()] = entry.Data()
	}
	return string(anchor.ToJSONFormatted(m))
}

func (t *trie) addNode(ctx *searchContext, node Node) error {
	if ctx.pointer == nil {
		t.root = newRootNode(t.digitizer.Base())
		ctx.pointer = t.root
	}

	entry := node.Value()
	for ctx.branchPosition < t.digitizer.NumDigitsOf(entry.Value())-1 {
		index, err := t.digitizer.DigitOf(entry.Value(), ctx.branchPosition)
		if err != nil {
			return err
		}

		childNode := newNode(t.digitizer.Base())
		if err := ctx.pointer.AddChild(index, childNode); err != nil {
			return err
		}
		ctx.pointer = childNode
		ctx.branchPosition++
	}

	index, err := t.digitizer.DigitOf(entry.Value(), ctx.branchPosition)
	if err != nil {
		return err
	}

	if err := ctx.pointer.AddChild(index, node); err != nil {
		return err
	}
	ctx.pointer = node
	ctx.branchPosition++
	return nil
}

func (t *trie) checkBounds(index int) error {
	if index < 0 || index >= t.Len() {
		return fmt.Errorf("trie: index out of bounds: Trie.Size() = %d, requested index = %d", t.Len(), index)
	}
	return nil
}

func (t *trie) find(ctx *searchContext, value string) (searchResult, error) {
	if value = strings.TrimSpace(value); value == "" {
		return -1, fmt.Errorf("trie: %w", hold.ErrNotFound)
	}

	if t.IsEmpty() {
		return Unmatched, nil
	}

	t.prepareSearch(ctx)

	numDigitsInElement := t.digitizer.NumDigitsOf(value)

	for ctx.pointer != nil && !ctx.atLeaf() {
		if ctx.branchPosition == numDigitsInElement {
			return Prefix, nil
		}

		m, err := ctx.descendTo(value)
		if err != nil {
			return -1, err
		}

		if m == childNotFound {
			return Unmatched, nil
		}
	}

	if ctx.pointer != nil && ctx.branchPosition != numDigitsInElement {
		return Extension, nil
	}
	return Matched, nil
}

func (t *trie) insert(entry Entry) (Node, error) {
	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	searchResult, err := t.find(ctx, entry.Value())
	if err != nil {
		return nil, err
	}

	if searchResult == Matched || (!t.digitizer.IsPrefixFree() && (searchResult == Prefix || searchResult == Extension)) {
		return nil, fmt.Errorf("trie: entry violates prefix-free requirement: %v", entry)
	}

	leaf := newLeaf()
	leaf.SetValue(entry)
	if err := t.addNode(ctx, leaf); err != nil {
		return nil, err
	}
	searchResult = Matched

	m, err := t.moveToPredecessor(ctx, entry.Value(), searchResult)
	if err != nil {
		return nil, err
	}

	if m {
		leaf.AddAfter(ctx.pointer.(Leaf))
	} else {
		leaf.AddAfter(t.head)
	}

	t.size++
	return leaf, nil
}

func (t *trie) moveToPredecessor(ctx *searchContext, value string, searchResult searchResult) (bool, error) {
	if ctx.atLeaf() && (searchResult == Greater || searchResult == Extension) {
		return true, nil
	}

	if searchResult != Greater {
		if err := ctx.retraceToLastLeftFork(value); err != nil {
			return false, err
		}
	}

	if ctx.atRoot() {
		return false, nil
	} else if !ctx.atLeaf() {
		ctx.moveToMaxDescendant()
	}
	return true, nil
}

func (t *trie) node(value string) (Node, error) {
	if t.IsEmpty() {
		return nil, fmt.Errorf("trie: %w", hold.ErrCollectionEmpty)
	}

	if value = strings.TrimSpace(value); value == "" {
		return nil, fmt.Errorf("trie: %w", hold.ErrValueRequired)
	}

	ctx := acquireSearchContext(t.digitizer)
	defer releaseSearchContext(ctx)

	r, err := t.find(ctx, value)
	if err != nil {
		return nil, err
	}

	if r == Matched {
		return ctx.pointer, nil
	}
	return nil, fmt.Errorf("trie: %w", hold.ErrNotFound)
}

func (t *trie) prepareSearch(ctx *searchContext) {
	ctx.digitizer = t.digitizer
	ctx.branchPosition = 0
	ctx.pointer = t.root
}

func (t *trie) remove(node Node) error {
	if leaf, ok := node.(Leaf); ok {
		leaf.Remove()
	}

	entry := node.Value()
	level := t.digitizer.NumDigitsOf(entry.Value())

	for !node.IsRoot() && !node.HasChildren() {
		parent := node.Parent()
		level--

		index, err := t.digitizer.DigitOf(entry.Value(), level)
		if err != nil {
			return err
		}

		parent.RemoveChildAt(index)
		node = parent
	}
	t.size--
	return nil
}
