package grid

func Area(vectors []Vector) int {
	coords := Travel(Coord{}, vectors)

	return Shoelace(coords) + Perimeter(coords)/2 + 1
}

func Travel(start Coord, vectors []Vector) []Coord {
	cur := start
	coords := make([]Coord, 0, len(vectors)+1)

	for i := range vectors {
		coords = append(coords, cur)
		cur = Add(cur, Mul(vectors[i].Dir, vectors[i].Len))
	}

	coords = append(coords, cur)

	return coords
}

func Shoelace(vertices []Coord) int {
	var n int

	for i := range vertices {
		next := (i + 1) % len(vertices)

		n += vertices[i].X * vertices[next].Y
		n -= vertices[i].Y * vertices[next].X
	}

	return Abs(n) / 2
}

func Perimeter(vertices []Coord) int {
	var n int

	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)

		sub := Sub(vertices[i], vertices[next])
		n += Abs(sub.X) + Abs(sub.Y)
	}

	return n
}
