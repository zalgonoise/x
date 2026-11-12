package mapping_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/x/mapping"
)

type faultyWriter[T any] struct {
	n   int
	err error
}

func (f faultyWriter[T]) Write(_ []T) (int, error) {
	return f.n, f.err
}

type faultyReader[T any] struct {
	items []T
	err   error
	done  bool
}

func (f *faultyReader[T]) Read(p []T) (int, error) {
	if f.done {
		return 0, f.err
	}

	f.done = true
	n := copy(p, f.items)

	return n, f.err
}

func TestTimeframeStreamGIO(t *testing.T) {
	i1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}
	i2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
	}
	i3 := mapping.Interval{
		From: time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	stream := mapping.NewTimeframeStream(mapping.WithFlatten(func(a, b int) int {
		return a + b
	}))

	// Write first two overlapping intervals via Write.
	n, err := stream.Write([]mapping.DataInterval[int]{
		{Interval: i1, Data: 10},
		{Interval: i2, Data: 20},
	})
	if err != nil || n != 2 {
		t.Fatalf("expected Write to return (2, nil), got (%d, %v)", n, err)
	}

	// Write third interval via WriteItem.
	if err := stream.WriteItem(mapping.DataInterval[int]{Interval: i3, Data: 5}); err != nil {
		t.Fatalf("expected WriteItem to succeed, got %v", err)
	}

	wants := []mapping.DataInterval[int]{
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			},
			Data: 10,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			},
			Data: 30,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
			},
			Data: 20,
		},
		{
			Interval: i3,
			Data:     5,
		},
	}

	// Zero-length Read should return (0, nil).
	if n, err := stream.Read(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on empty Read buffer, got (%d, %v)", n, err)
	}

	// ReadItem + UnreadItem.
	first, err := stream.ReadItem()
	if err != nil || first != wants[0] {
		t.Fatalf("expected first ReadItem %+v, got %+v (err=%v)", wants[0], first, err)
	}
	if err := stream.UnreadItem(); err != nil {
		t.Fatalf("expected UnreadItem to succeed, got %v", err)
	}
	if err := stream.UnreadItem(); !errors.Is(err, mapping.ErrUnreadAtStart) {
		t.Fatalf("expected ErrUnreadAtStart on second UnreadItem, got %v", err)
	}

	// ReadAll via gio.ReadAll.
	readAll, err := gio.ReadAll(stream)
	if err != nil {
		t.Fatalf("expected gio.ReadAll to succeed, got %v", err)
	}
	assertIntervalsEqual(t, wants, readAll)

	// Subsequent Read at EOF should return (0, io.EOF) and ReadItem should return io.EOF.
	buf := make([]mapping.DataInterval[int], 2)
	if n, err := stream.Read(buf); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected (0, io.EOF) at end of stream, got (%d, %v)", n, err)
	}
	if _, err := stream.ReadItem(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from ReadItem at end of stream, got %v", err)
	}

	// Underlying buffer is still intact for Slice() and All().
	assertIntervalsEqual(t, wants, stream.Slice())

	// ReadAt does not depend on or alter the read cursor.
	atBuf := make([]mapping.DataInterval[int], 2)
	if n, err := stream.ReadAt(atBuf, 1); err != nil || n != 2 {
		t.Fatalf("expected ReadAt(1) to return (2, nil), got (%d, %v)", n, err)
	}
	assertIntervalsEqual(t, wants[1:3], atBuf)

	// Partial ReadAt at end of buffer returns io.EOF.
	tailBuf := make([]mapping.DataInterval[int], 3)
	if n, err := stream.ReadAt(tailBuf, 2); !errors.Is(err, io.EOF) || n != 2 {
		t.Fatalf("expected ReadAt(2) to return (2, io.EOF), got (%d, %v)", n, err)
	}
	assertIntervalsEqual(t, wants[2:], tailBuf[:2])

	if _, err := stream.ReadAt(tailBuf, -1); !errors.Is(err, mapping.ErrNegativeOffset) {
		t.Fatalf("expected ErrNegativeOffset on negative offset, got %v", err)
	}
	if _, err := stream.ReadAt(tailBuf, 100); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on out-of-bounds ReadAt, got %v", err)
	}

	// Seek operations: SeekStart, SeekCurrent, SeekEnd.
	pos, err := stream.Seek(0, gio.SeekStart)
	if err != nil || pos != 0 {
		t.Fatalf("expected SeekStart(0) -> 0, got (%d, %v)", pos, err)
	}
	pos, err = stream.Seek(2, gio.SeekCurrent)
	if err != nil || pos != 2 {
		t.Fatalf("expected SeekCurrent(2) -> 2, got (%d, %v)", pos, err)
	}
	pos, err = stream.Seek(-1, gio.SeekEnd)
	if err != nil || pos != 3 {
		t.Fatalf("expected SeekEnd(-1) -> 3, got (%d, %v)", pos, err)
	}
	item, err := stream.ReadItem()
	if err != nil || item != wants[3] {
		t.Fatalf("expected last item %+v after SeekEnd(-1), got %+v (%v)", wants[3], item, err)
	}

	if _, err := stream.Seek(0, 99); !errors.Is(err, mapping.ErrInvalidWhence) {
		t.Fatalf("expected ErrInvalidWhence, got %v", err)
	}
	if _, err := stream.Seek(-10, gio.SeekStart); !errors.Is(err, mapping.ErrNegativePosition) {
		t.Fatalf("expected ErrNegativePosition, got %v", err)
	}

	// Reset clears buffer and resets cursor.
	stream.Reset()
	if stream.Len() != 0 {
		t.Fatalf("expected Len() == 0 after Reset, got %d", stream.Len())
	}
}

func TestTimeframeGIOAndCopyPipelines(t *testing.T) {
	i1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}
	i2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
	}

	srcTF := mapping.NewTimeframe[string]()
	if _, err := srcTF.Write([]mapping.DataInterval[string]{
		{Interval: i1, Data: "A"},
	}); err != nil {
		t.Fatalf("unexpected error on Timeframe.Write: %v", err)
	}
	if err := srcTF.WriteItem(mapping.DataInterval[string]{Interval: i2, Data: "B"}); err != nil {
		t.Fatalf("unexpected error on Timeframe.WriteItem: %v", err)
	}

	// Test Read, ReadItem, UnreadItem, ReadAt, Seek on Timeframe.
	if n, err := srcTF.Read(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on empty Read, got (%d, %v)", n, err)
	}
	first, err := srcTF.ReadItem()
	if err != nil || first.Data != "A" {
		t.Fatalf("expected first item A, got %+v (%v)", first, err)
	}
	if err := srcTF.UnreadItem(); err != nil {
		t.Fatalf("unexpected error on UnreadItem: %v", err)
	}
	if err := srcTF.UnreadItem(); !errors.Is(err, mapping.ErrUnreadAtStart) {
		t.Fatalf("expected ErrUnreadAtStart, got %v", err)
	}

	atBuf := make([]mapping.DataInterval[string], 1)
	if n, err := srcTF.ReadAt(atBuf, 1); err != nil || n != 1 || atBuf[0].Data != "B" {
		t.Fatalf("expected ReadAt(1) to return B, got %+v (%d, %v)", atBuf, n, err)
	}
	tailBuf := make([]mapping.DataInterval[string], 3)
	if n, err := srcTF.ReadAt(tailBuf, 1); !errors.Is(err, io.EOF) || n != 1 {
		t.Fatalf("expected ReadAt(1) with len 3 to return (1, io.EOF), got (%d, %v)", n, err)
	}
	if _, err := srcTF.ReadAt(tailBuf, -1); !errors.Is(err, mapping.ErrNegativeOffset) {
		t.Fatalf("expected ErrNegativeOffset, got %v", err)
	}
	if _, err := srcTF.ReadAt(tailBuf, 10); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}

	// Fan-out from srcTF to multiple TimeframeStreams with different strategies using gio.Copy + gio.MultiWriter.
	replaceStream := mapping.NewTimeframeStream(mapping.WithReplace[string]())
	keepStream := mapping.NewTimeframeStream(mapping.WithKeep[string]())
	flattenStream := mapping.NewTimeframeStream(mapping.WithFlatten(func(cur, next string) string {
		return cur + "+" + next
	}))

	multiWriter := gio.MultiWriter[mapping.DataInterval[string]](replaceStream, keepStream, flattenStream)
	copied, err := gio.Copy(multiWriter, srcTF)
	if err != nil || copied != 2 {
		t.Fatalf("expected gio.Copy to copy 2 items without error, got (%d, %v)", copied, err)
	}

	// At EOF on srcTF, Read and ReadItem return io.EOF, WriteTo returns (0, nil).
	if n, err := srcTF.Read(atBuf); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected (0, io.EOF), got (%d, %v)", n, err)
	}
	if _, err := srcTF.ReadItem(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF from ReadItem, got %v", err)
	}
	if n, err := srcTF.WriteTo(replaceStream); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) from WriteTo at EOF, got (%d, %v)", n, err)
	}

	assertIntervalsEqual(t, []mapping.DataInterval[string]{
		{Interval: mapping.Interval{From: i1.From, To: i2.From}, Data: "A"},
		{Interval: i2, Data: "B"},
	}, replaceStream.Slice())

	assertIntervalsEqual(t, []mapping.DataInterval[string]{
		{Interval: i1, Data: "A"},
		{Interval: mapping.Interval{From: i1.To, To: i2.To}, Data: "B"},
	}, keepStream.Slice())

	assertIntervalsEqual(t, []mapping.DataInterval[string]{
		{Interval: mapping.Interval{From: i1.From, To: i2.From}, Data: "A"},
		{Interval: mapping.Interval{From: i2.From, To: i1.To}, Data: "A+B"},
		{Interval: mapping.Interval{From: i1.To, To: i2.To}, Data: "B"},
	}, flattenStream.Slice())

	// Rewind srcTF with Seek and test ReadFrom on a new TimeframeStream and Timeframe.
	if pos, err := srcTF.Seek(0, gio.SeekCurrent); err != nil || pos != 2 {
		t.Fatalf("expected SeekCurrent(0) == 2, got (%d, %v)", pos, err)
	}
	if pos, err := srcTF.Seek(-2, gio.SeekEnd); err != nil || pos != 0 {
		t.Fatalf("expected SeekEnd(-2) == 0, got (%d, %v)", pos, err)
	}
	if _, err := srcTF.Seek(0, 99); !errors.Is(err, mapping.ErrInvalidWhence) {
		t.Fatalf("expected ErrInvalidWhence, got %v", err)
	}
	if _, err := srcTF.Seek(-1, gio.SeekStart); !errors.Is(err, mapping.ErrNegativePosition) {
		t.Fatalf("expected ErrNegativePosition, got %v", err)
	}

	dstStream := mapping.NewTimeframeStream[string]()
	nRead, err := dstStream.ReadFrom(gio.LimitReader[mapping.DataInterval[string]](srcTF, 2))
	if err != nil || nRead != 2 {
		t.Fatalf("expected ReadFrom to read 2 items, got (%d, %v)", nRead, err)
	}
	assertIntervalsEqual(t, replaceStream.Slice(), dstStream.Slice())

	// Test WriteTo from dstStream into a new Timeframe via ReadFrom / WriteTo.
	dstTF := mapping.NewTimeframe[string]()
	if _, err := srcTF.Seek(0, gio.SeekStart); err != nil {
		t.Fatalf("unexpected error on SeekStart: %v", err)
	}
	if nRead, err := dstTF.ReadFrom(gio.LimitReader[mapping.DataInterval[string]](srcTF, 2)); err != nil || nRead != 2 {
		t.Fatalf("expected Timeframe.ReadFrom to read 2 items, got (%d, %v)", nRead, err)
	}
	if nWritten, err := dstStream.WriteTo(dstTF); err != nil || nWritten != 2 {
		t.Fatalf("expected TimeframeStream.WriteTo to write 2 items, got (%d, %v)", nWritten, err)
	}
	if nWritten, err := dstStream.WriteTo(dstTF); err != nil || nWritten != 0 {
		t.Fatalf("expected second WriteTo at EOF to write 0 items, got (%d, %v)", nWritten, err)
	}

	dstTF.Reset()
	if dstTF.Len() != 0 {
		t.Fatalf("expected Len 0 after Reset, got %d", dstTF.Len())
	}
}

func TestTimeframeGIOErrorsAndNilReceivers(t *testing.T) {
	var (
		nilTF     *mapping.Timeframe[int]
		nilStream *mapping.TimeframeStream[int]
	)

	nilTF.Reset()
	nilStream.Reset()

	if _, err := nilTF.Write(nil); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilTF.Write, got %v", err)
	}
	if err := nilTF.WriteItem(mapping.DataInterval[int]{}); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilTF.WriteItem, got %v", err)
	}
	if _, err := nilTF.Read(make([]mapping.DataInterval[int], 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilTF.Read, got %v", err)
	}
	if _, err := nilTF.ReadAt(make([]mapping.DataInterval[int], 1), 0); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilTF.ReadAt, got %v", err)
	}
	if _, err := nilTF.ReadItem(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilTF.ReadItem, got %v", err)
	}
	if err := nilTF.UnreadItem(); !errors.Is(err, mapping.ErrUnreadAtStart) {
		t.Fatalf("expected ErrUnreadAtStart on nilTF.UnreadItem, got %v", err)
	}
	if _, err := nilTF.Seek(0, gio.SeekStart); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilTF.Seek, got %v", err)
	}
	if _, err := nilTF.ReadFrom(nil); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilTF.ReadFrom, got %v", err)
	}
	if n, err := nilTF.WriteTo(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on nilTF.WriteTo, got (%d, %v)", n, err)
	}

	if _, err := nilStream.Write(nil); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilStream.Write, got %v", err)
	}
	if err := nilStream.WriteItem(mapping.DataInterval[int]{}); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilStream.WriteItem, got %v", err)
	}
	if _, err := nilStream.Read(make([]mapping.DataInterval[int], 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilStream.Read, got %v", err)
	}
	if _, err := nilStream.ReadAt(make([]mapping.DataInterval[int], 1), 0); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilStream.ReadAt, got %v", err)
	}
	if _, err := nilStream.ReadItem(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF on nilStream.ReadItem, got %v", err)
	}
	if err := nilStream.UnreadItem(); !errors.Is(err, mapping.ErrUnreadAtStart) {
		t.Fatalf("expected ErrUnreadAtStart on nilStream.UnreadItem, got %v", err)
	}
	if _, err := nilStream.Seek(0, gio.SeekStart); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilStream.Seek, got %v", err)
	}
	if _, err := nilStream.ReadFrom(nil); !errors.Is(err, mapping.ErrNilReceiver) {
		t.Fatalf("expected ErrNilReceiver on nilStream.ReadFrom, got %v", err)
	}
	if n, err := nilStream.WriteTo(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on nilStream.WriteTo, got (%d, %v)", n, err)
	}

	// Test invalid intervals on Write, WriteItem, ReadFrom, and faulty Reader/Writer.
	valid := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}
	invalid := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}

	tf := mapping.NewTimeframe[int]()
	stream := mapping.NewTimeframeStream[int]()

	if n, err := tf.ReadFrom(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on tf.ReadFrom(nil), got (%d, %v)", n, err)
	}
	if n, err := stream.ReadFrom(nil); n != 0 || err != nil {
		t.Fatalf("expected (0, nil) on stream.ReadFrom(nil), got (%d, %v)", n, err)
	}

	if n, err := tf.Write([]mapping.DataInterval[int]{
		{Interval: valid, Data: 1},
		{Interval: invalid, Data: 2},
	}); n != 1 || !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected (1, ErrInvalidInterval) on tf.Write, got (%d, %v)", n, err)
	}
	if err := tf.WriteItem(mapping.DataInterval[int]{Interval: invalid, Data: 2}); !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval on tf.WriteItem, got %v", err)
	}

	if n, err := stream.Write([]mapping.DataInterval[int]{
		{Interval: valid, Data: 1},
		{Interval: invalid, Data: 2},
	}); n != 1 || !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected (1, ErrInvalidInterval) on stream.Write, got (%d, %v)", n, err)
	}
	if err := stream.WriteItem(mapping.DataInterval[int]{Interval: invalid, Data: 2}); !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval on stream.WriteItem, got %v", err)
	}

	// ReadFrom with invalid interval and ReadFrom with custom reader error.
	customErr := errors.New("read failure")
	if _, err := tf.ReadFrom(&faultyReader[mapping.DataInterval[int]]{
		items: []mapping.DataInterval[int]{{Interval: invalid, Data: 1}},
	}); !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval from tf.ReadFrom, got %v", err)
	}
	if _, err := tf.ReadFrom(&faultyReader[mapping.DataInterval[int]]{err: customErr}); !errors.Is(err, customErr) {
		t.Fatalf("expected customErr from tf.ReadFrom, got %v", err)
	}

	if _, err := stream.ReadFrom(&faultyReader[mapping.DataInterval[int]]{
		items: []mapping.DataInterval[int]{{Interval: invalid, Data: 1}},
	}); !errors.Is(err, mapping.ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval from stream.ReadFrom, got %v", err)
	}
	if _, err := stream.ReadFrom(&faultyReader[mapping.DataInterval[int]]{err: customErr}); !errors.Is(err, customErr) {
		t.Fatalf("expected customErr from stream.ReadFrom, got %v", err)
	}

	// WriteTo error cases: negative write count, short write, explicit write error.
	_, _ = tf.Seek(0, gio.SeekStart)
	if _, err := tf.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: -1}); !errors.Is(err, gio.ErrInvalidWrite) {
		t.Fatalf("expected gio.ErrInvalidWrite on tf.WriteTo, got %v", err)
	}
	if _, err := tf.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: 0, err: customErr}); !errors.Is(err, customErr) {
		t.Fatalf("expected customErr on tf.WriteTo, got %v", err)
	}
	if _, err := tf.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: 0, err: nil}); !errors.Is(err, gio.ErrShortWrite) {
		t.Fatalf("expected gio.ErrShortWrite on tf.WriteTo, got %v", err)
	}

	_, _ = stream.Seek(0, gio.SeekStart)
	if _, err := stream.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: -1}); !errors.Is(err, gio.ErrInvalidWrite) {
		t.Fatalf("expected gio.ErrInvalidWrite on stream.WriteTo, got %v", err)
	}
	if _, err := stream.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: 0, err: customErr}); !errors.Is(err, customErr) {
		t.Fatalf("expected customErr on stream.WriteTo, got %v", err)
	}
	if _, err := stream.WriteTo(faultyWriter[mapping.DataInterval[int]]{n: 0, err: nil}); !errors.Is(err, gio.ErrShortWrite) {
		t.Fatalf("expected gio.ErrShortWrite on stream.WriteTo, got %v", err)
	}
}
