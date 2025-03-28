package trie

import (
	"fmt"

	"github.com/transientvariable/collection-go"
)

type iterator struct {
	trie    *trie
	pointer Leaf
}

func newIterator(trie *trie, pointer Leaf) *iterator {
	return &iterator{trie: trie, pointer: pointer}
}

// HasNext ...
func (i *iterator) HasNext() bool {
	return i.hasNext()
}

// Next ...
func (i *iterator) Next() (string, error) {
	i.advance()
	entry, err := i.get()
	if err != nil {
		return "", err
	}
	return entry.Value(), nil
}

func (i *iterator) advance() bool {
	if i.pointer.IsTail() {
		return false
	}

	if !i.pointer.IsHead() && i.pointer.IsDeleted() {
		i.pointer = i.skipRemovedElements(i.pointer)
	} else {
		i.pointer = i.pointer.Next()
	}
	return !i.pointer.IsTail()
}

func (i *iterator) get() (Entry, error) {
	if !i.inCollection() {
		return nil, fmt.Errorf("trie: %w", collection.ErrNotFound)
	}
	return i.pointer.Value(), nil
}

func (i *iterator) hasNext() bool {
	if i.pointer.IsDeleted() {
		i.skipRemovedElements(i.pointer)
	}
	return !i.pointer.IsTail() && !i.pointer.Next().IsTail()
}

func (i *iterator) inCollection() bool {
	if i.pointer.IsHead() || i.pointer.IsTail() {
		return false
	}
	return !i.pointer.IsDeleted()
}

func (i *iterator) remove() error {
	if i.inCollection() {
		if err := i.trie.remove(i.pointer); err != nil {
			return err
		}
	}
	return nil
}

func (i *iterator) retreat() bool {
	if !i.pointer.IsTail() && !i.pointer.IsHead() && i.pointer.IsDeleted() {
		i.pointer = i.skipRemovedElements(i.pointer)
	}

	i.pointer = i.pointer.Previous()
	return !i.pointer.IsHead()
}

func (i *iterator) skipRemovedElements(leafNode Leaf) Leaf {
	if leafNode.IsHead() || leafNode.IsTail() || !leafNode.IsDeleted() {
		return leafNode
	}

	leafNode.SetNext(i.skipRemovedElements(leafNode.Next()))
	return leafNode.Next()
}
