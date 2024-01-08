package grid

import (
	"github.com/zalgonoise/gbuf"
)

const minAlloc = 64

type GraphType[T comparable] interface {
	Root() T
	Edges(T) []T
	IsLast(T) bool
}

type BFS[T comparable] struct {
	Queue    []T
	Cache    map[T]struct{}
	Previous map[T]T
	Len      map[T]uint64
}

func NewBFS[T comparable]() *BFS[T] {
	return &BFS[T]{
		Queue:    []T{},
		Cache:    make(map[T]struct{}),
		Previous: make(map[T]T),
		Len:      make(map[T]uint64),
	}
}

func (b *BFS[T]) Run(g GraphType[T]) T {
	root := g.Root()
	b.Queue = append(b.Queue, root)
	b.Len[root] = 0

	for len(b.Queue) > 0 {
		cur := b.Queue[0]
		b.Queue = b.Queue[1:]
		count := b.Len[cur]

		if _, ok := b.Cache[cur]; ok {
			continue
		}

		b.Cache[cur] = struct{}{}

		edges := g.Edges(cur)
		for i := range edges {
			if _, ok := b.Cache[edges[i]]; ok {
				continue
			}

			b.Len[edges[i]] = count + 1
			b.Previous[edges[i]] = cur

			if g.IsLast(edges[i]) {
				return edges[i]
			}

			b.Queue = append(b.Queue, edges[i])
		}
	}

	return root
}

type DFS[T comparable] struct {
	Queue []T
	Cache map[T]struct{}
	Len   map[T]uint64
}

func NewDFS[T comparable]() *DFS[T] {
	return &DFS[T]{
		Queue: make([]T, 0, minAlloc),
		Cache: make(map[T]struct{}),
		Len:   make(map[T]uint64),
	}
}

func (b *DFS[T]) Run(g GraphType[T]) T {
	root := g.Root()
	b.Queue = append(b.Queue, root)
	b.Len[root] = 0

	for len(b.Queue) > 0 {
		s := b.Queue[len(b.Queue)-1]
		b.Queue = b.Queue[:len(b.Queue)-1]
		d := b.Len[s]

		if _, ok := b.Cache[s]; ok {
			continue
		}

		b.Cache[s] = struct{}{}

		edges := g.Edges(s)
		for _, edge := range edges {
			if _, ok := b.Cache[edge]; ok {
				continue
			}

			b.Len[edge] = d + 1

			if g.IsLast(edge) {
				return edge
			}

			b.Queue = append(b.Queue, edge)
		}
	}

	return root
}

type Numeric interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}

type context[I Integer] struct {
	coord  Coord[I]
	dir    Coord[I]
	streak I
}

func AStar[T Numeric, S ~[]T, I Integer](g Graph[T, S, I], minSteps I, maxSteps I) I {
	var (
		startCtx = context[I]{
			coord:  g.Head,
			dir:    Coord[I]{},
			streak: 0,
		}
		pathQueue = &PriorityQueue[context[I], I]{}
		prevCtx   = make(map[context[I]]context[I])
		distance  = make(map[context[I]]I)
	)

	gbuf.Init[QueueItem[context[I], I]](pathQueue)
	gbuf.Push[QueueItem[context[I], I]](pathQueue, QueueItem[context[I], I]{
		Item:     startCtx,
		Priority: 0,
	})

	prevCtx[startCtx] = startCtx
	distance[startCtx] = 0

	for pathQueue.Len() > 0 {
		cur := gbuf.Pop[QueueItem[context[I], I]](pathQueue).(QueueItem[context[I], I]).Item
		curDistance := distance[cur]

		if g.IsLast(cur.coord) {
			return curDistance
		}

		edges := g.Edges(cur.coord)
		for i := range edges {
			dir := Sub(edges[i], cur.coord)
			streak := I(1)

			if dir == cur.dir {
				streak += cur.streak
			}

			nextCtx := context[I]{
				coord:  edges[i],
				dir:    dir,
				streak: streak,
			}

			nextDistance := curDistance + I(g.Map.Items[edges[i]])
			if length, ok := distance[nextCtx]; ok && nextDistance >= length {
				continue
			}

			if cur.streak < minSteps && dir != cur.dir && cur.coord != g.Head {
				continue
			}

			if streak > maxSteps {
				continue
			}

			if dir == Inverse(cur.dir) {
				continue
			}

			distance[nextCtx] = nextDistance
			prevCtx[nextCtx] = cur

			priority := nextDistance + Manhattan(edges[i], g.Tail)
			queueItem := QueueItem[context[I], I]{Item: nextCtx, Priority: priority}
			gbuf.Push[QueueItem[context[I], I]](pathQueue, queueItem)
		}
	}

	return -1
}
