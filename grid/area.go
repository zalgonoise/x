package grid

func Area[I Integer](vectors []Vector[I]) I {
	coords := Travel[I](Coord[I]{}, vectors)

	return Shoelace(coords) + Perimeter(coords)/2 + 1
}

func Travel[I Integer](start Coord[I], vectors []Vector[I]) []Coord[I] {
	cur := start
	coords := make([]Coord[I], 0, len(vectors)+1)

	for i := range vectors {
		coords = append(coords, cur)
		cur = Add(cur, Mul(vectors[i].Dir, I(vectors[i].Len)))
	}

	coords = append(coords, cur)

	return coords
}

func Shoelace[I Integer](vertices []Coord[I]) I {
	var n I

	for i := range vertices {
		next := (i + 1) % len(vertices)

		n += vertices[i].X * vertices[next].Y
		n -= vertices[i].Y * vertices[next].X
	}

	return Abs(n) / 2
}

func Perimeter[I Integer](vertices []Coord[I]) I {
	var n I

	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)

		sub := Sub(vertices[i], vertices[next])
		n += Abs(sub.X) + Abs(sub.Y)
	}

	return n
}
