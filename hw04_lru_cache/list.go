package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length    int
	firstItem *ListItem
	lastItem  *ListItem
}

func (l *list) Len() int {
	return l.length
}

func (l *list) Front() *ListItem {
	return l.firstItem
}

func (l *list) Back() *ListItem {
	return l.lastItem
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{
		Value: v,
		Next:  l.firstItem,
		Prev:  nil,
	}

	if l.firstItem != nil {
		l.firstItem.Prev = item
	}
	l.firstItem = item

	if l.lastItem == nil {
		l.lastItem = item
	}

	l.length++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{
		Value: v,
		Next:  nil,
		Prev:  l.lastItem,
	}

	if l.lastItem != nil {
		l.lastItem.Next = item
	}
	l.lastItem = item

	if l.firstItem == nil {
		l.firstItem = item
	}

	l.length++
	return item
}

func (l *list) Remove(i *ListItem) {
	if i == nil || l.length == 0 {
		return
	}

	if i == l.firstItem {
		l.firstItem = i.Next
		if l.firstItem != nil {
			l.firstItem.Prev = nil
		}
	} else {
		i.Prev.Next = i.Next
	}

	if i == l.lastItem {
		l.lastItem = i.Prev
		if l.lastItem != nil {
			l.lastItem.Next = nil
		}
	} else {
		i.Next.Prev = i.Prev
	}

	i.Prev = nil
	i.Next = nil
	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.length == 0 || i == l.firstItem {
		return
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	}

	if i == l.lastItem {
		l.lastItem = i.Prev
	}

	i.Prev = nil
	i.Next = l.firstItem

	if l.firstItem != nil {
		l.firstItem.Prev = i
	}
	l.firstItem = i

	if l.lastItem == nil {
		l.lastItem = i
	}
}

func NewList() List {
	return &list{}
}
