package grid

var (
	North = Coord{1, 0}
	South = Coord{-1, 0}
	East  = Coord{0, 1}
	West  = Coord{0, -1}
)

var Directions = []Coord{North, South, East, West}

type Vector struct {
	Dir Coord
	Len int
}

type Coord struct {
	Y int
	X int
}

func Add(a, b Coord) Coord {
	return Coord{
		Y: a.Y + b.Y,
		X: a.X + b.X,
	}
}

func Sub(a, b Coord) Coord {
	return Coord{
		Y: a.Y - b.Y,
		X: a.X - b.X,
	}
}

func Mul(c Coord, factor int) Coord {
	return Coord{c.Y * factor, c.X * factor}
}

func Abs(i int) int {
	if i < 0 {
		return -i
	}

	return i
}

func Inverse(c Coord) Coord {
	return Coord{
		Y: -c.Y,
		X: -c.X,
	}
}

func Manhattan(a, b Coord) int {
	return Abs(a.X-b.X) + Abs(a.Y-b.Y)
}
