package iter

// Seq defines a sequence iterator function for types K and T.
//
// The Seq will accept a yield function, which will consume the types K and T yielded from the iteration,
// and should return a true or false on whether the yielded data is OK and the sequence should continue.
//
// Finally, Seq returns a boolean on weather it finished successfully or not.
//
// This involves a two-way communication between Seq and its yield function, allowing eachother to continue
// once the former is completed, and vice-versa.
//
// Seq should be perceived as an idiomatic Go approach to a "for each" type of iterator.
type Seq[K any, T any] func(yield func(K, T) bool) bool
