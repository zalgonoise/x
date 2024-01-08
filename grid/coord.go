package grid

type Integer interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

func North[I Integer]() Coord[I] {
	return Coord[I]{Y: 1, X: 0}
}

func South[I Integer]() Coord[I] {
	return Coord[I]{Y: -1, X: 0}
}

func East[I Integer]() Coord[I] {
	return Coord[I]{Y: 0, X: 1}
}

func West[I Integer]() Coord[I] {
	return Coord[I]{Y: 0, X: -1}
}

func Directions[I Integer]() []Coord[I] {
	return []Coord[I]{North[I](), South[I](), East[I](), West[I]()}
}

type Vector[I Integer] struct {
	Dir Coord[I]
	Len int
}

type Coord[I Integer] struct {
	Y I
	X I
}

func Add[I Integer](a, b Coord[I]) Coord[I] {
	return Coord[I]{
		Y: a.Y + b.Y,
		X: a.X + b.X,
	}
}

func Sub[I Integer](a, b Coord[I]) Coord[I] {
	return Coord[I]{
		Y: a.Y - b.Y,
		X: a.X - b.X,
	}
}

func Mul[I Integer](c Coord[I], factor I) Coord[I] {
	return Coord[I]{c.Y * factor, c.X * factor}
}

func Abs[I Integer](i I) I {
	if i < 0 {
		return -i
	}

	return i
}

func Inverse[I Integer](c Coord[I]) Coord[I] {
	return Coord[I]{
		Y: -c.Y,
		X: -c.X,
	}
}

func Manhattan[I Integer](a, b Coord[I]) I {
	return Abs(a.X-b.X) + Abs(a.Y-b.Y)
}
