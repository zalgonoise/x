package grid

import (
	"github.com/zalgonoise/cfg"
)

type Quadrant uint8

const (
	Q1 Quadrant = iota // 0x00
	Q2                 // 0x01 - invert X axis
	Q3                 // 0x10 - invert Y axis
	Q4                 // 0x11 - invert both X and Y axis
)

type Map[T any, S ~[]T] struct {
	MaxY  int
	MaxX  int
	Items map[Coord]T
}

func Rebuild[T any, S ~[]T](m Map[T, S]) []S {
	var (
		typ  Quadrant
		maxX = m.MaxX
		maxY = m.MaxY
	)

	if m.MaxX < 0 {
		typ ^= Q2
		maxX = -maxX
	}

	if m.MaxY < 0 {
		typ ^= Q3
		maxY = -maxY
	}

	grid := make([]S, maxY+1)
	for i := range grid {
		grid[i] = make(S, maxX+1)
	}

	for coord, value := range m.Items {
		var x, y int

		switch typ {
		case Q1:
			y = maxY - coord.Y
			x = coord.X
		case Q2:
			x = coord.X + maxX
			y = maxY - coord.Y
		case Q3:
			x = coord.X
			y = -coord.Y
		case Q4:
			x = coord.X + maxX
			y = -coord.Y
		}

		grid[y][x] = value
	}

	return grid
}

func newQ1[T any, S ~[]T](items []S) Map[T, S] {
	maxY := len(items) - 1
	maxX := len(items[0]) - 1
	m := make(map[Coord]T)

	for y := range items {
		for x := range items[y] {
			m[Coord{Y: maxY - y, X: x}] = items[y][x]
		}
	}

	return Map[T, S]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,
	}
}

func newQ2[T any, S ~[]T](items []S) Map[T, S] {
	maxY := len(items) - 1
	maxX := -len(items[0]) + 1
	m := make(map[Coord]T)

	for y := range items {
		for x := range items[y] {
			m[Coord{Y: maxY - y, X: maxX + x}] = items[y][x]
		}
	}

	return Map[T, S]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,
	}
}

func newQ3[T any, S ~[]T](items []S) Map[T, S] {
	maxY := -len(items) + 1
	maxX := len(items[0]) - 1
	m := make(map[Coord]T)

	for y := range items {
		for x := range items[y] {
			m[Coord{Y: -y, X: x}] = items[y][x]
		}
	}

	return Map[T, S]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,
	}
}

func newQ4[T any, S ~[]T](items []S) Map[T, S] {
	maxY := -len(items) + 1
	maxX := -len(items[0]) + 1
	m := make(map[Coord]T)

	for y := range items {
		for x := range items[y] {
			m[Coord{Y: -y, X: maxX + x}] = items[y][x]
		}
	}

	return Map[T, S]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,
	}
}

func NewGrid[T any, S ~[]T](items []S, opts ...cfg.Option[MapConfig]) Map[T, S] {
	config := cfg.New(opts...)

	switch config.quadrant {
	case Q1:
		return newQ1[T, S](items)
	case Q2:
		return newQ2[T, S](items)
	case Q3:
		return newQ3[T, S](items)
	case Q4:
		return newQ4[T, S](items)
	default:
		return newQ1[T, S](items)
	}
}

type MapConfig struct {
	quadrant Quadrant
}

func WithQuadrant(quadrant Quadrant) cfg.Option[MapConfig] {
	return cfg.Register(func(config MapConfig) MapConfig {
		config.quadrant = quadrant

		return config
	})
}
