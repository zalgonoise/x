# `mapping`

Package `mapping` (`github.com/zalgonoise/x/mapping`) provides generic time-range (`Interval`) value mapping and conflict-resolution primitives for merging, splitting, and organizing values across overlapping timeframes.

---

## Overview

When working with time-bounded records (schedules, metrics, configuration overrides, feature flags, audit states), multiple intervals often overlap in time. `mapping` resolves overlapping intervals according to a configurable strategy and user-provided `MergeFunc[T]`, automatically splitting overlapping bounds into disjoint sub-intervals and coalescing adjacent intervals that carry equal values.

### How Overlapping Intervals Are Split

When an existing interval `A` (`[10:00, 14:00]`) and an incoming interval `B` (`[12:00, 16:00]`) overlap, the timeframe is partitioned into up to three regions:

```text
Existing (A):  |######### A #########|
Incoming (B):            |######### B #########|
               +---------+-----------+---------+
               10:00     12:00       14:00     16:00
               | Only A  |   A & B   | Only B  |
```

1. **`[10:00, 12:00)`** — Only `A` is active (`A.Data`).
2. **`[12:00, 14:00)`** — Both `A` and `B` overlap (`mergeFunc(A.Data, B.Data)`).
3. **`[14:00, 16:00)`** — Only `B` is active (`B.Data`).

After resolving the overlap, any adjacent sub-intervals whose values compare equal are automatically coalesced into a single continuous interval.

---

## Conflict-Resolution Strategies

Strategies are configured via functional options (`github.com/zalgonoise/cfg`):

| Option | Behavior on Overlap (`A & B`) | Result for `A [10:00, 14:00]` + `B [12:00, 16:00]` |
| :--- | :--- | :--- |
| `WithReplace[T]()` *(default)* | Incoming value `B` overwrites `A` (`MergeReplace`) | `[10:00, 12:00): A`, `[12:00, 16:00): B` |
| `WithKeep[T]()` | Existing value `A` is preserved (`MergeKeep`) | `[10:00, 14:00): A`, `[14:00, 16:00): B` |
| `WithFlatten[T](mergeFunc)` / `WithMergeFunc[T](mergeFunc)` | Combines values via `mergeFunc(cur, next)` | `[10:00, 12:00): A`, `[12:00, 14:00): merge(A, B)`, `[14:00, 16:00): B` |
| `WithIntersect[T](mergeFunc)` | Retains only intersecting sub-intervals | `[12:00, 14:00): merge(A, B)` |
| `WithAppend[T]()` | Appends and sorts chronologically by `(From, To)` without splitting | `[10:00, 14:00): A`, `[12:00, 16:00): B` |

### Additional Options

- **`WithCompareFunc[T](cmp func(a, b T) bool)`**: Custom equality check used to coalesce adjacent intervals with equal data. For Go `comparable` types, `==` is used by default. Required when `T` is non-comparable (e.g., `maps.Equal` for `map[K]V`) if adjacent coalescing is desired.
- **`WithOffset[T](offset time.Duration)`**: Configures a discrete boundary offset between split intervals (e.g., `time.Second` for inclusive timestamp ranges `[12:00:00, 13:59:59]` and `[14:00:00, 16:00:00]`). Defaults to `0` (half-open/contiguous intervals `[From, To)`).

---

## Usage

### 1. Stateless Slice & Iterator Functions (`Organize` / `OrganizeSeq`)

```go
package main

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/mapping"
)

func main() {
	items := []mapping.DataInterval[int]{
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			},
			Data: 10,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
			},
			Data: 25,
		},
	}

	organized := mapping.Organize(
		items,
		mapping.WithFlatten(func(cur, next int) int {
			return cur + next
		}),
	)

	for _, item := range organized {
		fmt.Printf("%s -> %s: %d\n",
			item.Interval.From.Format("15:04"),
			item.Interval.To.Format("15:04"),
			item.Data,
		)
	}
	// Output:
	// 10:00 -> 12:00: 10
	// 12:00 -> 14:00: 35
	// 14:00 -> 16:00: 25
}
```

### 2. Batch Container (`Timeframe[T]`)

`Timeframe[T]` buffers intervals in $O(1)$ amortized time as they are added and resolves them in batch when `Organize` or `Merge` is called:

```go
tf := mapping.NewTimeframe[string]()
tf.Add(interval1, "primary")
tf.Add(interval2, "override")

// Organize using the default Replace strategy (or pass strategy options to Override)
organized := tf.Organize(mapping.WithReplace[string]())

for interval, val := range organized.All() {
	// Iterate over organized (Interval, T) pairs using standard library iter.Seq2
	_ = interval
	_ = val
}
```

### 3. Streaming Container (`TimeframeStream[T]`)

`TimeframeStream[T]` resolves overlaps eagerly on every `Add`, `AddSlice`, or `Append` call using binary search ($O(\log N)$ lookup) to maintain a continuously sorted, disjoint slice:

```go
stream := mapping.NewTimeframeStream(
	mapping.WithKeep[string](),
)

stream.Add(interval1, "first")
stream.Add(interval2, "second")

for interval, val := range stream.All() {
	_ = interval
	_ = val
}
```

### 4. Generic I/O Streams (`github.com/zalgonoise/gio`)

Both `TimeframeStream[T]` and `Timeframe[T]` implement the generic I/O interfaces from [`github.com/zalgonoise/gio`](https://github.com/zalgonoise/gio) over `DataInterval[T]`:

- `gio.Reader[DataInterval[T]]` / `gio.Writer[DataInterval[T]]` / `gio.ReadWriter[DataInterval[T]]`
- `gio.Seeker[DataInterval[T]]` / `gio.ReadWriteSeeker[DataInterval[T]]`
- `gio.ReaderAt[DataInterval[T]]`
- `gio.ItemReader[DataInterval[T]]` / `gio.ItemScanner[DataInterval[T]]` / `gio.ItemWriter[DataInterval[T]]`
- `gio.ReaderFrom[DataInterval[T]]` / `gio.WriterTo[DataInterval[T]]`

`Read`, `ReadItem`, `UnreadItem`, `Seek`, and `WriteTo` operate on a non-destructive read cursor over the container's buffer (`All()` and `Slice()` always return the full underlying sequence, and `Seek(0, io.SeekStart)` rewinds the cursor). `Write`, `WriteItem`, and `ReadFrom` validate each incoming `DataInterval[T]` and stop at the first invalid interval, returning `(n, ErrInvalidInterval)`.

```go
// Copy unorganized items from a batch Timeframe into a TimeframeStream,
// resolving overlaps on the fly, or fan out with gio.MultiWriter:
raw := mapping.NewTimeframe[int]()
_ = raw.Add(interval1, 10)
_ = raw.Add(interval2, 20)

sumStream := mapping.NewTimeframeStream(
	mapping.WithFlatten(func(cur, next int) int { return cur + next }),
)

_, err := gio.Copy(sumStream, raw)
if err != nil {
	panic(err)
}

resolved, err := gio.ReadAll(sumStream)
```

