package trie

import (
	"fmt"

	"github.com/pkg/errors"
)

// Node ...
type Node interface {
	AddChild(index int, child Node) error
	ChildAt(index int) (Node, error)
	Children() []Node
	HasChildren() bool
	IsLeaf() bool
	IsRoot() bool
	Parent() Node
	SetParent(parent Node)
	SetValue(entry Entry)
	RemoveChildAt(index int) bool
	Value() Entry
}

type node struct {
	children    []Node
	isRoot      bool
	numChildren int
	parent      Node
	value       Entry
}

func newNode(capacity int) Node {
	if capacity <= 0 {
		return &node{}
	}
	return &node{children: make([]Node, capacity)}
}

func newRootNode(capacity int) Node {
	return &node{
		children: make([]Node, capacity),
		isRoot:   true,
	}
}

// AddChild ...
func (n *node) AddChild(index int, child Node) error {
	if index < 0 || index >= len(n.children) {
		return errors.Errorf("trie: index out of bounds for node: capacity = %d, requested index = %d", cap(n.children), index)
	}

	if n.children[index] != nil {
		return errors.Errorf("child exists at index %v", index)
	}

	if n.children[index] == nil {
		n.numChildren++
	}

	n.children[index] = child
	child.SetParent(n)
	return nil
}

// ChildAt ...
func (n *node) ChildAt(index int) (Node, error) {
	if err := n.checkBounds(index); err != nil {
		return nil, err
	}
	return n.children[index], nil
}

// Children ...
func (n *node) Children() []Node {
	return n.children
}

// HasChildren ...
func (n *node) HasChildren() bool {
	return n.numChildren > 0
}

// IsLeaf ...
func (n *node) IsLeaf() bool {
	return false
}

// IsRoot ...
func (n *node) IsRoot() bool {
	return n.isRoot
}

// Parent ...
func (n *node) Parent() Node {
	return n.parent
}

// RemoveChildAt ...
func (n *node) RemoveChildAt(index int) bool {
	if err := n.checkBounds(index); err != nil {
		return false
	}

	if n.children[index] != nil {
		n.children[index] = nil
		n.numChildren--
		return true
	}
	return false
}

// SetParent ...
func (n *node) SetParent(parent Node) {
	n.parent = parent
}

// SetValue ...
func (n *node) SetValue(entry Entry) {
	n.value = entry
}

// Value ...
func (n *node) Value() Entry {
	return n.value
}

// String ...
func (n *node) String() string {
	return fmt.Sprintf("%v", n.value)
}

func (n *node) checkBounds(index int) error {
	if index < 0 || index > len(n.children) {
		return errors.Errorf("index out of bounds [Node.capacity = %v, requested index = %v]", cap(n.children), index)
	}
	return nil
}

// Leaf ...
type Leaf interface {
	Node

	AddAfter(leafNode Leaf)
	IsDeleted() bool
	IsHead() bool
	IsTail() bool
	Next() Leaf
	Previous() Leaf
	SetNext(next Leaf)
	SetPrevious(previous Leaf)
	Remove()
}

type leaf struct {
	next     Leaf
	node     Node
	previous Leaf
	isHead   bool
	isTail   bool
}

// AddChild delegates the call to Node.AddChild for the Leaf.
func (l *leaf) AddChild(index int, child Node) error {
	return l.node.AddChild(index, child)
}

// ChildAt delegates the call to Node.ChildAt for the Leaf.
func (l *leaf) ChildAt(index int) (Node, error) {
	return l.node.ChildAt(index)
}

// Children delegates the call to Node.Children for the Leaf.
func (l *leaf) Children() []Node {
	return l.node.Children()
}

// HasChildren delegates the call to Node.HasChildren for the Leaf.
func (l *leaf) HasChildren() bool {
	return l.node.HasChildren()
}

// IsRoot delegates the call to Node.IsRoot for the Leaf.
func (l *leaf) IsRoot() bool {
	return l.node.IsRoot()
}

// Parent delegates the call to Node.Parent for the Leaf.
func (l *leaf) Parent() Node {
	return l.node.Parent()
}

// SetParent delegates the call to Node.SetParent for the Leaf.
func (l *leaf) SetParent(parent Node) {
	l.node.SetParent(parent)
}

// SetValue delegates the call to Node.SetValue for the Leaf.
func (l *leaf) SetValue(entry Entry) {
	l.node.SetValue(entry)
}

// RemoveChildAt delegates the call to Node.RemoveChildAt for the Leaf.
func (l *leaf) RemoveChildAt(index int) bool {
	return l.node.RemoveChildAt(index)
}

// Value delegates the call to Node.Value for the Leaf.
func (l *leaf) Value() Entry {
	return l.node.Value()
}

func newLeaf() Leaf {
	return &leaf{node: newNode(0)}
}

// AddAfter ...
func (l *leaf) AddAfter(leafNode Leaf) {
	l.SetNext(leafNode.Next())
	leafNode.SetNext(l)
	l.SetPrevious(leafNode)
	l.next.SetPrevious(l)
}

// IsDeleted ...
func (l *leaf) IsDeleted() bool {
	return l.previous == nil
}

// IsHead ...
func (l *leaf) IsHead() bool {
	return l.isHead
}

// IsLeaf ...
func (l *leaf) IsLeaf() bool {
	return true
}

// IsTail ...
func (l *leaf) IsTail() bool {
	return l.isTail
}

// Next ...
func (l *leaf) Next() Leaf {
	return l.next
}

// Previous ...
func (l *leaf) Previous() Leaf {
	return l.previous
}

// SetNext ...
func (l *leaf) SetNext(next Leaf) {
	l.next = next
}

// SetPrevious ...
func (l *leaf) SetPrevious(previous Leaf) {
	l.previous = previous
}

// Remove ...
func (l *leaf) Remove() {
	l.previous.SetNext(l.next)
	l.next.SetPrevious(l.previous)
	l.markDeleted()
}

func (l *leaf) markDeleted() {
	l.previous = nil
}
