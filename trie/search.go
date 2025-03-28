package trie

import (
	"reflect"
	"sync"

	"github.com/transientvariable/collection-go"
)

type searchResult int

const (
	Extension searchResult = iota + 1
	Greater
	Less
	Matched
	Prefix
	Unmatched
)

const childNotFound = -1

var searchContextPool = sync.Pool{
	New: func() any { return &searchContext{} },
}

func acquireSearchContext(digitizer Digitizer) *searchContext {
	ctx := searchContextPool.Get().(*searchContext)
	ctx.digitizer = digitizer
	return ctx
}

func releaseSearchContext(ctx *searchContext) {
	ctx.pointer = nil
	ctx.digitizer = nil
	ctx.branchPosition = 0
	ctx.numMatches = 0
	searchContextPool.Put(ctx)
}

type searchContext struct {
	pointer        Node
	digitizer      Digitizer
	branchPosition int
	numMatches     int
}

func (s *searchContext) ascend() int {
	s.branchPosition -= 1
	s.pointer = s.pointer.Parent()
	return s.branchPosition
}

func (s *searchContext) atLeaf() bool {
	if _, ok := s.pointer.(Leaf); ok {
		return true
	}
	return false
}

func (s *searchContext) atRoot() bool {
	return s.pointer.Parent() == nil
}

func (s *searchContext) childIndexOf(value string) (int, error) {
	return s.digitizer.DigitOf(value, s.branchPosition)
}

func (s *searchContext) descendTo(value string) (int, error) {
	index, err := s.digitizer.DigitOf(value, s.branchPosition)
	if err != nil {
		return -1, err
	}
	return s.descendToIndex(index), nil
}

func (s *searchContext) descendToIndex(index int) int {
	child, err := s.pointer.ChildAt(index)
	if err != nil || child == nil {
		return childNotFound
	}

	s.branchPosition += 1
	s.pointer = child
	return index
}

func (s *searchContext) entriesInSubtree(collection collection.Collection[string]) error {
	if s.atLeaf() {
		if err := collection.Add(s.pointer.Value().Value()); err != nil {
			return err
		}
		return nil
	}

	for i := 0; i < s.digitizer.Base(); i++ {
		if s.descendToIndex(i) != childNotFound {
			if err := s.entriesInSubtree(collection); err != nil {
				return err
			}
			s.ascend()
		}
	}
	return nil
}

func (s *searchContext) extendPath(value string, node Node) (int, error) {
	index, err := s.digitizer.DigitOf(value, s.branchPosition)
	if err != nil {
		return -1, err
	}

	if err := node.AddChild(index, node); err != nil {
		return -1, err
	}
	return s.descendToIndex(index), nil
}

func (s *searchContext) moveToMaxDescendant() {
	for !s.atLeaf() {
		index := s.digitizer.Base() - 1
		for s.descendToIndex(index) == childNotFound {
			index--
		}
	}
}

// TODO: method argument still needed?
func (s *searchContext) processedEndOfString(_ string) (bool, error) {
	childNode, err := s.pointer.Parent().ChildAt(0)
	if err != nil {
		return false, err
	}
	return s.digitizer.IsPrefixFree() &&
		!s.pointer.IsRoot() &&
		reflect.DeepEqual(childNode, s.pointer), nil
}

func (s *searchContext) retraceToLastLeftFork(value string) error {
	for {
		if !s.atLeaf() {
			index, err := s.digitizer.DigitOf(value, s.branchPosition)
			if err != nil {
				return err
			}

			for i := index - 1; i >= 0; i-- {
				if s.descendToIndex(i) != childNotFound {
					return nil
				}
			}
		}

		if s.atRoot() {
			return nil
		}
		s.ascend()
	}
}
