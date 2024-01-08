package grid

import (
	"testing"

	"github.com/zalgonoise/cfg"
)

func TestNewGraph(t *testing.T) {
	data := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	mapQ1 := NewGrid[int](data, WithQuadrant(Q1))
	mapQ2 := NewGrid[int](data, WithQuadrant(Q2))
	mapQ3 := NewGrid[int](data, WithQuadrant(Q3))
	mapQ4 := NewGrid[int](data, WithQuadrant(Q4))

	for _, testcase := range []struct {
		name  string
		input Map[int, []int, int]
		head  *Coord[int]
		tail  *Coord[int]
		wants Graph[int, []int, int]
	}{
		{
			name:  "Q1/NoOpts",
			input: mapQ1,
			wants: Graph[int, []int, int]{
				Head: Coord[int]{0, 0},
				Tail: Coord[int]{3, 3},
				Map:  mapQ1,
			},
		},
		{
			name:  "Q2/NoOpts",
			input: mapQ2,
			wants: Graph[int, []int, int]{
				Head: Coord[int]{0, 0},
				Tail: Coord[int]{3, -3},
				Map:  mapQ2,
			},
		},
		{
			name:  "Q3/NoOpts",
			input: mapQ3,
			wants: Graph[int, []int, int]{
				Head: Coord[int]{0, 0},
				Tail: Coord[int]{-3, 3},
				Map:  mapQ3,
			},
		},
		{
			name:  "Q4/NoOpts",
			input: mapQ4,
			wants: Graph[int, []int, int]{
				Head: Coord[int]{0, 0},
				Tail: Coord[int]{-3, -3},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithHead",
			input: mapQ4,
			head:  &Coord[int]{-1, -1},
			wants: Graph[int, []int, int]{
				Head: Coord[int]{-1, -1},
				Tail: Coord[int]{-3, -3},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithTail",
			input: mapQ4,
			tail:  &Coord[int]{-1, -1},
			wants: Graph[int, []int, int]{
				Head: Coord[int]{0, 0},
				Tail: Coord[int]{-1, -1},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithHeadAndTail",
			input: mapQ4,
			head:  &Coord[int]{-1, -1},
			tail:  &Coord[int]{-2, -2},
			wants: Graph[int, []int, int]{
				Head: Coord[int]{-1, -1},
				Tail: Coord[int]{-2, -2},
				Map:  mapQ4,
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			opts := make([]cfg.Option[GraphConfig[int]], 0, 2)

			if testcase.head != nil {
				opts = append(opts, WithRoot(*testcase.head))
			}

			if testcase.tail != nil {
				opts = append(opts, WithEnd(*testcase.tail))
			}

			g := NewGraph(testcase.input, opts...)

			isEqual(t, testcase.wants.Head, g.Head)
			isEqual(t, testcase.wants.Tail, g.Tail)
		})
	}
}
