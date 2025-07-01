package collections

import "container/heap"

type QueItem[T any] struct {
	Value    T
	Priority int
}

func NewPriorityQue[T any]() *PriorityQue[T] {
	return &PriorityQue[T]{
		que:         make([]*QueItem[T], 0),
		initialized: false,
	}
}

type PriorityQue[T any] struct {
	que         []*QueItem[T]
	initialized bool
}

func (q *PriorityQue[T]) Len() int {
	return len(q.que)
}

func (q *PriorityQue[T]) Less(i, j int) bool {
	return q.que[i].Priority < q.que[j].Priority
}

func (q *PriorityQue[T]) Swap(i, j int) {
	q.que[i], q.que[j] = q.que[j], q.que[i]
}

func (q *PriorityQue[T]) Push(x any) {
	item, ok := x.(*QueItem[T])
	if !ok {
		panic("PriorityQue.Push: x is not a *QueItem[T]")
	}
	q.que = append(q.que, item)
}

func (q *PriorityQue[T]) Pop() any {
	old, n := q.que, len(q.que)
	item := old[n-1]
	q.que = old[0 : n-1]
	return item
}

func (q *PriorityQue[T]) InitHeap() {
	heap.Init(q)
}

func (q *PriorityQue[T]) HeapPush(item *QueItem[T]) {
	heap.Push(q, item)
}

func (q *PriorityQue[T]) HeapPop() *QueItem[T] {
	el, ok := heap.Pop(q).(*QueItem[T])
	if !ok {
		panic("popped element is not *QueItem[T]")
	}
	return el
}
