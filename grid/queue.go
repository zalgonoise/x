package grid

type QueueItem[T any, I Integer] struct {
	Item     T
	Priority I
}

type PriorityQueue[T any, I Integer] struct {
	q []QueueItem[T, I]
}

func (q *PriorityQueue[T, I]) Len() int {
	return len(q.q)
}

func (q *PriorityQueue[T, I]) Less(i, j int) bool {
	return q.q[i].Priority < q.q[j].Priority
}

func (q *PriorityQueue[T, I]) Swap(i, j int) {
	q.q[i], q.q[j] = q.q[j], q.q[i]
}

func (q *PriorityQueue[T, I]) Push(x QueueItem[T, I]) {
	q.q = append(q.q, x)
}

func (q *PriorityQueue[T, I]) Pop() QueueItem[T, I] {
	old := q.q
	n := len(old)

	item := old[n-1]
	q.q = old[0 : n-1]

	return item
}
