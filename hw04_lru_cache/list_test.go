package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
	t.Run("empty list with Remove", func(t *testing.T) {
		l := NewList()
		l.Remove(nil)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		l.Remove(&ListItem{})
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
	t.Run("empty list with MoveToFront", func(t *testing.T) {
		l := NewList()
		l.MoveToFront(nil)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		l.MoveToFront(&ListItem{})
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
	t.Run("list has one element with MoveToFront", func(t *testing.T) {
		l := NewList()
		i := l.PushFront(nil)
		l.MoveToFront(i)
		require.Equal(t, 1, l.Len())

		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Front().Next)
		require.Nil(t, l.Front().Value)

		require.Nil(t, l.Back().Value)
		require.Nil(t, l.Back().Prev)
		require.Nil(t, l.Back().Next)

		require.NotNil(t, l.Front())
		require.NotNil(t, l.Back())
	})
	t.Run("list has one element with Remove", func(t *testing.T) {
		l := NewList()
		i := l.PushFront(nil)
		l.Remove(i)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
	t.Run("list has six elements with MoveToFront", func(t *testing.T) {
		l := NewList()
		l.PushFront(10) // [10]
		l.PushFront(20) // [10, 20]
		l.PushFront(30) // [10, 20, 30]
		l.PushFront(40)
		l.PushFront(50)
		l.PushFront(60)
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{60, 50, 40, 30, 20, 10}, elems)
		l.MoveToFront(l.Back())
		l.MoveToFront(l.Back())
		l.MoveToFront(l.Back())

		elems = make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{30, 20, 10, 60, 50, 40}, elems)
	})
	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}
