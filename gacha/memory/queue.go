package memory

import (
	"container/list"
	"time"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

const notQueued = -1

// record is an outbox or dead-letter entry. Dead letters use updatedAt as their failure time.
type record[R any] struct {
	id          uuid.UUID
	data        R
	status      gacha.Status
	attempts    int
	lastError   string
	createdAt   time.Time
	updatedAt   time.Time
	availableAt time.Time

	// index is the record's position in the claim queue, or notQueued.
	index int
	// elem is the record's element in the completed or dead-letter list, or nil.
	elem *list.Element
}

func (r *record[R]) entry() gacha.Entry[R] {
	return gacha.Entry[R]{
		ID:        r.id,
		Data:      r.data,
		Attempts:  r.attempts,
		CreatedAt: r.createdAt,
	}
}

// queue is a min-heap of the claimable (pending and processing) records, ordered by availability time, then ID. It
// implements heap.Interface.
type queue[R any] []*record[R]

func (q queue[R]) Len() int {
	return len(q)
}

func (q queue[R]) Less(i, j int) bool {
	if !q[i].availableAt.Equal(q[j].availableAt) {
		return q[i].availableAt.Before(q[j].availableAt)
	}

	return q[i].id.Compare(q[j].id) < 0
}

func (q queue[R]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue[R]) Push(x any) {
	rec, _ := x.(*record[R])
	rec.index = len(*q)
	*q = append(*q, rec)
}

func (q *queue[R]) Pop() any {
	old := *q
	n := len(old) - 1
	rec := old[n]
	old[n] = nil
	rec.index = notQueued
	*q = old[:n]

	return rec
}
