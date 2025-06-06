package heap

import "cmp"

// Sort sorts an array of integers in ascending order using the heap sort algorithm.
func Sort[T cmp.Ordered](arr []T) {
	n := len(arr)

	// build a max heap from the input data.
	// the loop starts from the last non-leaf node and goes up to the root.
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	// one by one extract elements from the heap.
	for i := n - 1; i > 0; i-- {
		// move the current root (maximum element) to the end.
		arr[0], arr[i] = arr[i], arr[0]

		// call max heapify on the reduced heap.
		heapify(arr, i, 0)
	}
}

// heapify ensures that the subtree rooted at index i is a max heap.
// n is the size of the heap.
func heapify[T cmp.Ordered](arr []T, n int, i int) {
	largest := i     // initialize largest as root
	left := 2*i + 1  // left child
	right := 2*i + 2 // right child

	// if the left child is larger than the root.
	if left < n && arr[left] > arr[largest] {
		largest = left
	}

	// if the right child is larger than the largest so far.
	if right < n && arr[right] > arr[largest] {
		largest = right
	}

	// if the largest element is not the root.
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i] // swap

		// recursively heapify the affected subtree.
		heapify(arr, n, largest)
	}
}
