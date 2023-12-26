package grid

import "github.com/zalgonoise/cfg"

type Graph[T any, S ~[]T] struct {
	Head Coord
	Tail Coord

	Map Map[T, S]
}

func (g Graph[T, S]) Root() Coord {
	return g.Head
}

func (g Graph[T, S]) Edges(c Coord) []Coord {
	edges := make([]Coord, 0, 4)

	north := Add(c, North)
	south := Add(c, South)
	east := Add(c, East)
	west := Add(c, West)

	if _, ok := g.Map.Items[north]; ok {
		edges = append(edges, north)
	}
	if _, ok := g.Map.Items[south]; ok {
		edges = append(edges, south)
	}
	if _, ok := g.Map.Items[east]; ok {
		edges = append(edges, east)
	}
	if _, ok := g.Map.Items[west]; ok {
		edges = append(edges, west)
	}

	return edges
}

func (g Graph[T, S]) IsLast(c Coord) bool {
	return c == g.Tail
}

func NewGraph[T any, S ~[]T](m Map[T, S], opts ...cfg.Option[GraphConfig]) Graph[T, S] {
	config := cfg.New(opts...)

	var tail Coord

	switch config.tail {
	case nil:
		tail = Coord{m.MaxY, m.MaxX}
	default:
		tail = *config.tail
	}

	return Graph[T, S]{
		Head: config.head,
		Tail: tail,
		Map:  m,
	}
}

type GraphConfig struct {
	head Coord
	tail *Coord
}

func WithRoot(c Coord) cfg.Option[GraphConfig] {
	return cfg.Register(func(config GraphConfig) GraphConfig {
		config.head = c

		return config
	})
}

func WithEnd(c Coord) cfg.Option[GraphConfig] {
	return cfg.Register(func(config GraphConfig) GraphConfig {
		config.tail = &c

		return config
	})
}
