package grid

import (
	"fmt"
	"testing"
)

func TestNewGrid(t *testing.T) {
	data := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	for _, testcase := range []struct {
		name     string
		input    [][]int
		quadrant Quadrant
		wants    Map[int, []int]
	}{
		{
			name:     "Q1",
			input:    data,
			quadrant: Q1,
			wants: Map[int, []int]{
				MaxX: 3,
				MaxY: 3,
				Items: map[Coord]int{
					{3, 0}: 1,
					{3, 1}: 2,
					{3, 2}: 3,
					{3, 3}: 4,
					{2, 0}: 5,
					{2, 1}: 6,
					{2, 2}: 7,
					{2, 3}: 8,
					{1, 0}: 9,
					{1, 1}: 10,
					{1, 2}: 11,
					{1, 3}: 12,
					{0, 0}: 13,
					{0, 1}: 14,
					{0, 2}: 15,
					{0, 3}: 16,
				},
			},
		},
		{
			name:     "Q2",
			input:    data,
			quadrant: Q2,
			wants: Map[int, []int]{
				MaxX: -3,
				MaxY: 3,
				Items: map[Coord]int{
					{3, -3}: 1,
					{3, -2}: 2,
					{3, -1}: 3,
					{3, 0}:  4,
					{2, -3}: 5,
					{2, -2}: 6,
					{2, -1}: 7,
					{2, 0}:  8,
					{1, -3}: 9,
					{1, -2}: 10,
					{1, -1}: 11,
					{1, 0}:  12,
					{0, -3}: 13,
					{0, -2}: 14,
					{0, -1}: 15,
					{0, 0}:  16,
				},
			},
		},
		{
			name:     "Q3",
			input:    data,
			quadrant: Q3,
			wants: Map[int, []int]{
				MaxX: 3,
				MaxY: -3,
				Items: map[Coord]int{
					{0, 0}:  1,
					{0, 1}:  2,
					{0, 2}:  3,
					{0, 3}:  4,
					{-1, 0}: 5,
					{-1, 1}: 6,
					{-1, 2}: 7,
					{-1, 3}: 8,
					{-2, 0}: 9,
					{-2, 1}: 10,
					{-2, 2}: 11,
					{-2, 3}: 12,
					{-3, 0}: 13,
					{-3, 1}: 14,
					{-3, 2}: 15,
					{-3, 3}: 16,
				},
			},
		},
		{
			name:     "Q4",
			input:    data,
			quadrant: Q4,
			wants: Map[int, []int]{
				MaxX: -3,
				MaxY: -3,
				Items: map[Coord]int{
					{0, -3}:  1,
					{0, -2}:  2,
					{0, -1}:  3,
					{0, 0}:   4,
					{-1, -3}: 5,
					{-1, -2}: 6,
					{-1, -1}: 7,
					{-1, 0}:  8,
					{-2, -3}: 9,
					{-2, -2}: 10,
					{-2, -1}: 11,
					{-2, 0}:  12,
					{-3, -3}: 13,
					{-3, -2}: 14,
					{-3, -1}: 15,
					{-3, 0}:  16,
				},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			g := NewGrid(testcase.input, WithQuadrant(testcase.quadrant))

			isEqual(t, testcase.wants.MaxX, g.MaxX)
			isEqual(t, testcase.wants.MaxY, g.MaxY)

			for key, value := range testcase.wants.Items {
				v, ok := g.Items[key]
				isEqual(t, true, ok)
				isEqual(t, value, v)
			}
		})
	}
}

func TestRebuild(t *testing.T) {
	data := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	for _, quadrant := range []Quadrant{Q1, Q2, Q3, Q4} {
		t.Run(fmt.Sprintf("%d", quadrant+1), func(t *testing.T) {
			for _, testcase := range []struct {
				name  string
				input [][]int
			}{
				{
					name:  "Simple",
					input: data,
				},
			} {
				{
					t.Run(testcase.name, func(t *testing.T) {
						grid := NewGrid(testcase.input, WithQuadrant(quadrant))

						rebuilt := Rebuild(grid)

						t.Log(rebuilt)

						isEqual(t, len(testcase.input), len(rebuilt))
						for i := range testcase.input {
							isEqual(t, len(testcase.input[i]), len(rebuilt[i]))
							for idx := range testcase.input[i] {
								isEqual(t, testcase.input[i][idx], rebuilt[i][idx])
							}
						}
					})
				}
			}
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
		{
			name:  "Q1/WayOffGrid/Inverted",
			input: mapQ1,
			key:   Coord{-10, -11},
			wants: 6,
		},
		{
			name:  "Q2/WayOffGrid/Inverted",
			input: mapQ2,
			key:   Coord{-10, 11},
			wants: 7,
		},
		{
			name:  "Q3/WayOffGrid/Inverted",
			input: mapQ3,
			key:   Coord{10, -11},
			wants: 10,
		},
		{
			name:  "Q4/WayOffGrid/Inverted",
			input: mapQ4,
			key:   Coord{10, 11},
			wants: 11,
		},
		{
			name:  "Q1/WayOffGrid/Alike",
			input: mapQ1,
			key:   Coord{7, 7},
			wants: 10,
		},
		{
			name:  "Q2/WayOffGrid/Alike",
			input: mapQ2,
			key:   Coord{7, -7},
			wants: 11,
		},
		{
			name:  "Q3/WayOffGrid/Alike",
			input: mapQ3,
			key:   Coord{-7, 7},
			wants: 6,
		},
		{
			name:  "Q4/WayOffGrid/Alike",
			input: mapQ4,
			key:   Coord{-7, -7},
			wants: 7,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			value, ok := Get(testcase.input, testcase.key)
			isEqual(t, true, ok)
			isEqual(t, testcase.wants, value)
		})
	}
}

func isEqual[T comparable](t *testing.T, wants, got T) {
	if got != wants {
		t.Errorf("output mismatch error: wanted %v ; got %v", wants, got)
		t.Fail()

		return
	}

	t.Logf("output matched expected value: %v", wants)
}
