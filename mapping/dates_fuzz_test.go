package mapping

import (
	"testing"
	"time"
)

func FuzzReplace(f *testing.F) {
	f.Add(
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC).Unix(),
	)

	f.Fuzz(func(t *testing.T, aFrom, aTo, bFrom, bTo int64) {
		interval1 := Interval{
			From: time.Unix(aFrom, 0),
			To:   time.Unix(aTo, 0),
		}
		interval2 := Interval{
			From: time.Unix(bFrom, 0),
			To:   time.Unix(bTo, 0),
		}

		_, _ = replace(interval1, interval2, 0)
	})
}

func FuzzSplit(f *testing.F) {
	f.Add(
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC).Unix(),
	)

	f.Fuzz(func(t *testing.T, aFrom, aTo, bFrom, bTo int64) {
		interval1 := Interval{
			From: time.Unix(aFrom, 0),
			To:   time.Unix(aTo, 0),
		}
		interval2 := Interval{
			From: time.Unix(bFrom, 0),
			To:   time.Unix(bTo, 0),
		}

		_, _ = split(interval1, interval2, 0)
	})
}
