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

	mapQ1 := NewGrid(data, WithQuadrant(Q1))
	mapQ2 := NewGrid(data, WithQuadrant(Q2))
	mapQ3 := NewGrid(data, WithQuadrant(Q3))
	mapQ4 := NewGrid(data, WithQuadrant(Q4))

	for _, testcase := range []struct {
		name  string
		input Map[int, []int]
		head  *Coord
		tail  *Coord
		wants Graph[int, []int]
	}{
		{
			name:  "Q1/NoOpts",
			input: mapQ1,
			wants: Graph[int, []int]{
				Head: Coord{0, 0},
				Tail: Coord{3, 3},
				Map:  mapQ1,
			},
		},
		{
			name:  "Q2/NoOpts",
			input: mapQ2,
			wants: Graph[int, []int]{
				Head: Coord{0, 0},
				Tail: Coord{3, -3},
				Map:  mapQ2,
			},
		},
		{
			name:  "Q3/NoOpts",
			input: mapQ3,
			wants: Graph[int, []int]{
				Head: Coord{0, 0},
				Tail: Coord{-3, 3},
				Map:  mapQ3,
			},
		},
		{
			name:  "Q4/NoOpts",
			input: mapQ4,
			wants: Graph[int, []int]{
				Head: Coord{0, 0},
				Tail: Coord{-3, -3},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithHead",
			input: mapQ4,
			head:  &Coord{-1, -1},
			wants: Graph[int, []int]{
				Head: Coord{-1, -1},
				Tail: Coord{-3, -3},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithTail",
			input: mapQ4,
			tail:  &Coord{-1, -1},
			wants: Graph[int, []int]{
				Head: Coord{0, 0},
				Tail: Coord{-1, -1},
				Map:  mapQ4,
			},
		},
		{
			name:  "Q4/WithHeadAndTail",
			input: mapQ4,
			head:  &Coord{-1, -1},
			tail:  &Coord{-2, -2},
			wants: Graph[int, []int]{
				Head: Coord{-1, -1},
				Tail: Coord{-2, -2},
				Map:  mapQ4,
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			opts := make([]cfg.Option[GraphConfig], 0, 2)

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

func TestGet(t *testing.T) {
	data := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	mapQ1 := NewGrid(data, WithQuadrant(Q1), WithReflection())
	mapQ2 := NewGrid(data, WithQuadrant(Q2), WithReflection())
	mapQ3 := NewGrid(data, WithQuadrant(Q3), WithReflection())
	mapQ4 := NewGrid(data, WithQuadrant(Q4), WithReflection())

	for _, testcase := range []struct {
		name  string
		input Map[int, []int]
		key   Coord
		wants int
	}{
		{
			name:  "Q1/WithinGrid",
			input: mapQ1,
			key:   Coord{0, 0},
			wants: 13,
		},
		{
			name:  "Q1/OffGrid/Y",
			input: mapQ1,
			key:   Coord{-1, 0},
			wants: 9,
		},
		{
			name:  "Q1/OffGrid/X",
			input: mapQ1,
			key:   Coord{1, -1},
			wants: 10,
		},
		{
			name:  "Q1/OffGrid/YX",
			input: mapQ1,
			key:   Coord{-2, -1},
			wants: 6,
		},
		{
			name:  "Q2/WithinGrid",
			input: mapQ2,
			key:   Coord{0, 0},
			wants: 16,
		},
		{
			name:  "Q2/OffGrid/Y",
			input: mapQ2,
			key:   Coord{-1, 0},
			wants: 12,
		},
		{
			name:  "Q2/OffGrid/X",
			input: mapQ2,
			key:   Coord{1, 1},
			wants: 11,
		},
		{
			name:  "Q2/OffGrid/YX",
			input: mapQ2,
			key:   Coord{-2, -1},
			wants: 7,
		},
		{
			name:  "Q3/WithinGrid",
			input: mapQ3,
			key:   Coord{0, 0},
			wants: 1,
		},
		{
			name:  "Q3/OffGrid/Y",
			input: mapQ3,
			key:   Coord{1, 0},
			wants: 5,
		},
		{
			name:  "Q3/OffGrid/X",
			input: mapQ3,
			key:   Coord{-1, -1},
			wants: 6,
		},
		{
			name:  "Q3/OffGrid/YX",
			input: mapQ3,
			key:   Coord{2, -1},
			wants: 10,
		},

		{
			name:  "Q4/WithinGrid",
			input: mapQ4,
			key:   Coord{0, 0},
			wants: 4,
		},
		{
			name:  "Q4/OffGrid/Y",
			input: mapQ4,
			key:   Coord{1, 0},
			wants: 8,
		},
		{
			name:  "Q4/OffGrid/X",
			input: mapQ4,
			key:   Coord{-1, 1},
			wants: 7,
		},
		{
			name:  "Q4/OffGrid/YX",
			input: mapQ4,
			key:   Coord{2, 1},
			wants: 11,
		},
		//TODO: add way-off-grid tests with the correct resolution logic in place
		// currently it's not reflecting the grid correctly as it's not accounting for the inversions appropriately
	} {
		t.Run(testcase.name, func(t *testing.T) {
			value, ok := Get(testcase.input, testcase.key)
			isEqual(t, true, ok)
			isEqual(t, testcase.wants, value)
		})
	}
}
