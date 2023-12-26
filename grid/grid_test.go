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

func isEqual[T comparable](t *testing.T, wants, got T) {
	if got != wants {
		t.Errorf("output mismatch error: wanted %v ; got %v", wants, got)
		t.Fail()

		return
	}

	t.Logf("output matched expected value: %v", wants)
}
