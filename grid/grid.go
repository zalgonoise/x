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

type Map[T any, S ~[]T, I Integer] struct {
	MaxY  I
	MaxX  I
	Items map[Coord[I]]T

	reflects bool
}

func Get[T any, S ~[]T, I Integer](m Map[T, S, I], coord Coord[I]) (T, bool) {
	data, ok := m.Items[coord]

	if ok {
		return data, true
	}

	var zero T

	if !ok && !m.reflects {
		return zero, false
	}

	var (
		typ  Quadrant
		maxX = m.MaxX
		maxY = m.MaxY
	)

	if coord.Y < 0 {
		coord.Y = -coord.Y
	}

	if coord.X < 0 {
		coord.X = -coord.X
	}

	if m.MaxX < 0 {
		typ ^= Q2
		maxX = -maxX
	}

	if m.MaxY < 0 {
		typ ^= Q3
		maxY = -maxY
	}

	newCoord := Coord[I]{
		Y: I(int(coord.Y) % int(m.MaxY)),
		X: I(int(coord.X) % int(m.MaxX)),
	}

	switch typ {
	case Q1:
		if int(coord.Y/m.MaxY)%2 == 1 {
			newCoord.Y = m.MaxY - newCoord.Y
		}

		if int(coord.X/m.MaxX)%2 == 1 {
			newCoord.X = m.MaxX - newCoord.X
		}

		v, ok := m.Items[newCoord]

		return v, ok
	case Q2:
		if int(coord.Y/Abs(m.MaxY))%2 == 1 {
			newCoord.Y = m.MaxY - newCoord.Y
		}

		if int(coord.X/Abs(m.MaxX))%2 == 1 {
			newCoord.X = m.MaxX + Abs(newCoord.X)
		}

		if newCoord.X > 0 {
			newCoord.X = -newCoord.X
		}

		v, ok := m.Items[newCoord]

		return v, ok
	case Q3:
		if int(coord.Y/Abs(m.MaxY))%2 == 1 {
			newCoord.Y = m.MaxY + Abs(newCoord.Y)
		}

		if int(coord.X/Abs(m.MaxX))%2 == 1 {
			newCoord.X = m.MaxX - newCoord.X
		}

		if newCoord.Y > 0 {
			newCoord.Y = -newCoord.Y
		}

		v, ok := m.Items[newCoord]

		return v, ok
	case Q4:
		if int(coord.Y/Abs(m.MaxY))%2 == 1 {
			newCoord.Y = m.MaxY + Abs(newCoord.Y)
		}

		if int(coord.X/Abs(m.MaxX))%2 == 1 {
			newCoord.X = m.MaxX + Abs(newCoord.X)
		}

		if newCoord.Y > 0 {
			newCoord.Y = -newCoord.Y
		}

		if newCoord.X > 0 {
			newCoord.X = -newCoord.X
		}

		v, ok := m.Items[newCoord]

		return v, ok
	}

	return zero, false
}

func Rebuild[T any, S ~[]T, I Integer](m Map[T, S, I]) []S {
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

	grid := make([]S, int(maxY+1))
	for i := range grid {
		grid[i] = make(S, int(maxX+1))
	}

	for coord, value := range m.Items {
		var x, y I

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

		grid[int(y)][int(x)] = value
	}

	return grid
}

func newQ1[T any, S ~[]T, I Integer](items []S, reflects bool) Map[T, S, I] {
	maxY := I(len(items) - 1)
	maxX := I(len(items[0]) - 1)
	m := make(map[Coord[I]]T)

	for y := range items {
		for x := range items[y] {
			m[Coord[I]{Y: maxY - I(y), X: I(x)}] = items[y][x]
		}
	}

	return Map[T, S, I]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,

		reflects: reflects,
	}
}

func newQ2[T any, S ~[]T, I Integer](items []S, reflects bool) Map[T, S, I] {
	maxY := I(len(items) - 1)
	maxX := I(-len(items[0]) + 1)
	m := make(map[Coord[I]]T)

	for y := range items {
		for x := range items[y] {
			m[Coord[I]{Y: maxY - I(y), X: maxX + I(x)}] = items[y][x]
		}
	}

	return Map[T, S, I]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,

		reflects: reflects,
	}
}

func newQ3[T any, S ~[]T, I Integer](items []S, reflects bool) Map[T, S, I] {
	maxY := I(-len(items) + 1)
	maxX := I(len(items[0]) - 1)
	m := make(map[Coord[I]]T)

	for y := range items {
		for x := range items[y] {
			m[Coord[I]{Y: I(-y), X: I(x)}] = items[y][x]
		}
	}

	return Map[T, S, I]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,

		reflects: reflects,
	}
}

func newQ4[T any, S ~[]T, I Integer](items []S, reflects bool) Map[T, S, I] {
	maxY := I(-len(items) + 1)
	maxX := I(-len(items[0]) + 1)
	m := make(map[Coord[I]]T)

	for y := range items {
		for x := range items[y] {
			m[Coord[I]{Y: I(-y), X: maxX + I(x)}] = items[y][x]
		}
	}

	return Map[T, S, I]{
		MaxY:  maxY,
		MaxX:  maxX,
		Items: m,

		reflects: reflects,
	}
}

func NewGrid[I Integer, T any, S ~[]T](items []S, opts ...cfg.Option[MapConfig]) Map[T, S, I] {
	config := cfg.New(opts...)

	switch config.quadrant {
	case Q1:
		return newQ1[T, S, I](items, config.reflection)
	case Q2:
		return newQ2[T, S, I](items, config.reflection)
	case Q3:
		return newQ3[T, S, I](items, config.reflection)
	case Q4:
		return newQ4[T, S, I](items, config.reflection)
	default:
		return newQ1[T, S, I](items, config.reflection)
	}
}

type MapConfig struct {
	quadrant   Quadrant
	reflection bool
}

func WithQuadrant(quadrant Quadrant) cfg.Option[MapConfig] {
	return cfg.Register(func(config MapConfig) MapConfig {
		config.quadrant = quadrant

		return config
	})
}

func WithReflection() cfg.Option[MapConfig] {
	return cfg.Register(func(config MapConfig) MapConfig {
		config.reflection = true

		return config
	})
}
