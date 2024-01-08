package grid

import "github.com/zalgonoise/cfg"

type Graph[T any, S ~[]T, I Integer] struct {
	Head Coord[I]
	Tail Coord[I]

	Map Map[T, S, I]
}

func (g Graph[T, S, I]) Root() Coord[I] {
	return g.Head
}

func (g Graph[T, S, I]) Edges(c Coord[I]) []Coord[I] {
	edges := make([]Coord[I], 0, 4)

	north := Add(c, North[I]())
	south := Add(c, South[I]())
	east := Add(c, East[I]())
	west := Add(c, West[I]())

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

func (g Graph[T, S, I]) IsLast(c Coord[I]) bool {
	return c == g.Tail
}

func NewGraph[T any, S ~[]T, I Integer](m Map[T, S, I], opts ...cfg.Option[GraphConfig[I]]) Graph[T, S, I] {
	config := cfg.New(opts...)

	var tail Coord[I]

	switch config.tail {
	case nil:
		tail = Coord[I]{m.MaxY, m.MaxX}
	default:
		tail = *config.tail
	}

	return Graph[T, S, I]{
		Head: config.head,
		Tail: tail,
		Map:  m,
	}
}

type GraphConfig[I Integer] struct {
	head Coord[I]
	tail *Coord[I]
}

func WithRoot[I Integer](c Coord[I]) cfg.Option[GraphConfig[I]] {
	return cfg.Register(func(config GraphConfig[I]) GraphConfig[I] {
		config.head = c

		return config
	})
}

func WithEnd[I Integer](c Coord[I]) cfg.Option[GraphConfig[I]] {
	return cfg.Register(func(config GraphConfig[I]) GraphConfig[I] {
		config.tail = &c

		return config
	})
}
