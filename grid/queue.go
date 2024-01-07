package grid

type QueueItem[T any] struct {
	Item     T
	Priority int
}

type PriorityQueue[T any] struct {
	q []QueueItem[T]
}

func (q *PriorityQueue[T]) Len() int {
	return len(q.q)
}

func (q *PriorityQueue[T]) Less(i, j int) bool {
	return q.q[i].Priority < q.q[j].Priority
}

func (q *PriorityQueue[T]) Swap(i, j int) {
	q.q[i], q.q[j] = q.q[j], q.q[i]
}

func (q *PriorityQueue[T]) Push(x QueueItem[T]) {
	q.q = append(q.q, x)
}

func (q *PriorityQueue[T]) Pop() QueueItem[T] {
	old := q.q
	n := len(old)

	item := old[n-1]
	q.q = old[0 : n-1]

	return item
}
