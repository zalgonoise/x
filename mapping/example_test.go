package mapping_test

import (
	"fmt"
	"time"

	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/x/mapping"
)

func ExampleOrganize_replace() {
	i1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
	}
	i2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	items := []mapping.DataInterval[string]{
		{Interval: i1, Data: "primary"},
		{Interval: i2, Data: "override"},
	}

	organized := mapping.Organize(items, mapping.WithReplace[string]())
	for _, item := range organized {
		fmt.Printf("%s -> %s: %s\n",
			item.Interval.From.Format("15:04"),
			item.Interval.To.Format("15:04"),
			item.Data,
		)
	}
	// Output:
	// 10:00 -> 12:00: primary
	// 12:00 -> 14:00: override
	// 14:00 -> 16:00: primary
}

func ExampleOrganize_flatten() {
	i1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}
	i2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
	}

	items := []mapping.DataInterval[int]{
		{Interval: i1, Data: 10},
		{Interval: i2, Data: 25},
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

func ExampleTimeframe() {
	tf := mapping.NewTimeframe(
		mapping.WithFlatten(func(cur, next string) string {
			return cur + "+" + next
		}),
	)

	tf.Add(mapping.Interval{
		From: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}, "A")

	tf.Add(mapping.Interval{
		From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}, "B")

	for interval, val := range tf.Organize().All() {
		fmt.Printf("%s -> %s: %s\n",
			interval.From.Format("15:04"),
			interval.To.Format("15:04"),
			val,
		)
	}
	// Output:
	// 09:00 -> 11:00: A
	// 11:00 -> 12:00: A+B
	// 12:00 -> 14:00: B
}

func ExampleTimeframeStream() {
	stream := mapping.NewTimeframeStream(
		mapping.WithKeep[string](),
	)

	stream.Add(mapping.Interval{
		From: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}, "first")

	stream.Add(mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
	}, "second")

	for interval, val := range stream.All() {
		fmt.Printf("%s -> %s: %s\n",
			interval.From.Format("15:04"),
			interval.To.Format("15:04"),
			val,
		)
	}
	// Output:
	// 09:00 -> 12:00: first
	// 12:00 -> 15:00: second
}

func ExampleTimeframeStream_gioCopy() {
	src := mapping.NewTimeframe[int]()
	_, _ = src.Write([]mapping.DataInterval[int]{
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
			Data: 20,
		},
	})

	stream := mapping.NewTimeframeStream(
		mapping.WithFlatten(func(cur, next int) int {
			return cur + next
		}),
	)

	_, _ = gio.Copy(stream, src)

	items, _ := gio.ReadAll(stream)
	for _, item := range items {
		fmt.Printf("%s -> %s: %d\n",
			item.Interval.From.Format("15:04"),
			item.Interval.To.Format("15:04"),
			item.Data,
		)
	}
	// Output:
	// 10:00 -> 12:00: 10
	// 12:00 -> 14:00: 30
	// 14:00 -> 16:00: 20
}
