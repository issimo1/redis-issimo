package list

import "errors"

type Consumer func(i int, val interface{}) bool
type Expected func(actual interface{}) bool

type LinkedList struct {
	head *node
	last *node
	size int
}

type node struct {
	prev *node
	next *node
	val  any
}

func newNode(val any) *node {
	return &node{val: val}
}

func (l *LinkedList) Add(val interface{}) {
	n := newNode(val)
	if l.last == nil {
		l.head = n
		l.last = n
	} else {
		n.prev = l.last
		l.last.next = n
		l.last = n
	}
	l.size++
}

func (l *LinkedList) find(idx int) *node {
	if idx < l.Len()/2 {
		n := l.head
		for i := 0; i < idx; i++ {
			n = n.next
		}
		return n
	}
	n := l.last
	for i := l.Len() - 1; i > idx; i-- {
		n = n.prev
	}
	return n
}

func (l *LinkedList) Len() int {
	return l.size
}

func (l *LinkedList) Get(idx int) (any, error) {
	if idx < 0 || idx >= l.Len() {
		return nil, errors.New("out of range")
	}
	return l.find(idx).val, nil
}

func (l *LinkedList) Modify(idx int, val any) error {
	if idx < 0 || idx >= l.Len() {
		return errors.New("out of range")
	}
	l.find(idx).val = val
	return nil
}

func (l *LinkedList) delNode(n *node) {
	prev := n.prev
	next := n.next

	if prev != nil {
		prev.next = next
	} else {
		l.head = next
	}

	if next != nil {
		next.prev = prev
	} else {
		l.last = next
	}
	n.prev = nil
	n.next = nil
	return
}

func (l *LinkedList) Del(idx int) (any, error) {
	if idx < 0 || idx >= l.Len() {
		return nil, errors.New("out of range")
	}
	n := l.find(idx)
	l.delNode(n)
	return n.val, nil
}

func (l *LinkedList) DelLastNode() (any, error) {
	if l.Len() == 0 {
		return nil, nil
	}
	return l.Del(l.Len() - 1)
}

func (l *LinkedList) ForEach(consumer Consumer) {
	i := 0
	for n := l.head; n != nil; n = n.next {
		if !consumer(i, n.val) {
			break
		}
	}
}

func (l *LinkedList) Contains(expect Expected) bool {
	result := false
	l.ForEach(func(idx int, val interface{}) bool {
		if expect(val) {
			result = true
			return false
		}
		return true
	})
	return result
}

func (l *LinkedList) DelAllByVal(expected Expected) int {
	removed := 0
	for n := l.head; n != nil; {
		next := n.next
		if expected(n.val) {
			l.delNode(n)
			removed++
		}
		n = next
	}
	return removed
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}
