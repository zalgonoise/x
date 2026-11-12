package mapping

import (
	"errors"
	"io"

	"github.com/zalgonoise/gio"
)

// I/O errors returned by Timeframe and TimeframeStream gio methods.
var (
	ErrInvalidInterval  = errors.New("mapping: invalid interval")
	ErrNilReceiver      = errors.New("mapping: nil receiver")
	ErrUnreadAtStart    = errors.New("mapping: cannot unread item at beginning of buffer")
	ErrInvalidWhence    = errors.New("mapping: invalid seek whence")
	ErrNegativePosition = errors.New("mapping: negative seek position")
	ErrNegativeOffset   = errors.New("mapping: negative read offset")
)

var (
	_ gio.ReadWriter[DataInterval[any]]      = (*TimeframeStream[any])(nil)
	_ gio.ReadWriteSeeker[DataInterval[any]] = (*TimeframeStream[any])(nil)
	_ gio.ReaderAt[DataInterval[any]]        = (*TimeframeStream[any])(nil)
	_ gio.ItemScanner[DataInterval[any]]     = (*TimeframeStream[any])(nil)
	_ gio.ItemWriter[DataInterval[any]]      = (*TimeframeStream[any])(nil)
	_ gio.ReaderFrom[DataInterval[any]]      = (*TimeframeStream[any])(nil)
	_ gio.WriterTo[DataInterval[any]]        = (*TimeframeStream[any])(nil)

	_ gio.ReadWriter[DataInterval[any]]      = (*Timeframe[any])(nil)
	_ gio.ReadWriteSeeker[DataInterval[any]] = (*Timeframe[any])(nil)
	_ gio.ReaderAt[DataInterval[any]]        = (*Timeframe[any])(nil)
	_ gio.ItemScanner[DataInterval[any]]     = (*Timeframe[any])(nil)
	_ gio.ItemWriter[DataInterval[any]]      = (*Timeframe[any])(nil)
	_ gio.ReaderFrom[DataInterval[any]]      = (*Timeframe[any])(nil)
	_ gio.WriterTo[DataInterval[any]]        = (*Timeframe[any])(nil)
)

// Reset clears all buffered intervals and resets the read offset to 0 on the
// [Timeframe], retaining the underlying storage capacity for future writes.
func (t *Timeframe[T]) Reset() {
	if t == nil {
		return
	}

	clear(t.buffer)
	t.buffer = t.buffer[:0]
	t.readIdx = 0
}

// Write appends each [DataInterval] in p to the [Timeframe] buffer,
// implementing [gio.Writer].
//
// If any element in p carries an invalid [Interval], Write stops at that
// element and returns the number of elements written so far along with
// [ErrInvalidInterval].
func (t *Timeframe[T]) Write(p []DataInterval[T]) (n int, err error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	t.init()

	for i := range p {
		if !t.Add(p[i].Interval, p[i].Data) {
			return i, ErrInvalidInterval
		}
	}

	return len(p), nil
}

// WriteItem appends a single [DataInterval] to the [Timeframe] buffer,
// implementing [gio.ItemWriter]. It returns [ErrInvalidInterval] if item's
// interval is invalid.
func (t *Timeframe[T]) WriteItem(item DataInterval[T]) error {
	if t == nil {
		return ErrNilReceiver
	}

	if !t.Add(item.Interval, item.Data) {
		return ErrInvalidInterval
	}

	return nil
}

// Read copies up to len(p) unread [DataInterval] items from the [Timeframe]
// into p, advancing the internal read cursor and implementing [gio.Reader].
//
// The underlying buffer is retained so [Timeframe.All], [Timeframe.Slice],
// [Timeframe.Organize], and [Timeframe.Seek] continue to access all intervals.
func (t *Timeframe[T]) Read(p []DataInterval[T]) (n int, err error) {
	if t == nil {
		return 0, io.EOF
	}

	if len(p) == 0 {
		return 0, nil
	}

	if t.readIdx >= int64(len(t.buffer)) {
		return 0, io.EOF
	}

	n = copy(p, t.buffer[t.readIdx:])
	t.readIdx += int64(n)

	return n, nil
}

// ReadAt copies up to len(p) [DataInterval] items starting at offset off in the
// [Timeframe] buffer without modifying the read cursor, implementing [gio.ReaderAt].
func (t *Timeframe[T]) ReadAt(p []DataInterval[T], off int64) (n int, err error) {
	if off < 0 {
		return 0, ErrNegativeOffset
	}

	if t == nil || off >= int64(len(t.buffer)) {
		return 0, io.EOF
	}

	n = copy(p, t.buffer[off:])
	if n < len(p) {
		err = io.EOF
	}

	return n, err
}

// ReadItem reads and returns the next unread [DataInterval] from the [Timeframe],
// advancing the read cursor by 1 and implementing [gio.ItemReader].
func (t *Timeframe[T]) ReadItem() (DataInterval[T], error) {
	if t == nil || t.readIdx >= int64(len(t.buffer)) {
		var zero DataInterval[T]
		return zero, io.EOF
	}

	item := t.buffer[t.readIdx]
	t.readIdx++

	return item, nil
}

// UnreadItem steps the read cursor back by one position, implementing
// [gio.ItemScanner]. It returns [ErrUnreadAtStart] if the cursor is already at
// index 0.
func (t *Timeframe[T]) UnreadItem() error {
	if t == nil || t.readIdx <= 0 {
		return ErrUnreadAtStart
	}

	t.readIdx--

	return nil
}

// Seek sets the offset for the next Read or WriteTo operation on the [Timeframe],
// implementing [gio.Seeker].
func (t *Timeframe[T]) Seek(offset int64, whence int) (int64, error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	var abs int64

	switch whence {
	case gio.SeekStart:
		abs = offset
	case gio.SeekCurrent:
		abs = t.readIdx + offset
	case gio.SeekEnd:
		abs = int64(len(t.buffer)) + offset
	default:
		return 0, ErrInvalidWhence
	}

	if abs < 0 {
		return 0, ErrNegativePosition
	}

	t.readIdx = abs

	return abs, nil
}

// ReadFrom reads [DataInterval] items from r until EOF or error and appends
// them to the [Timeframe], implementing [gio.ReaderFrom].
func (t *Timeframe[T]) ReadFrom(r gio.Reader[DataInterval[T]]) (n int64, err error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	if r == nil {
		return 0, nil
	}

	buf := make([]DataInterval[T], minAlloc)

	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := t.Write(buf[:nr])
			n += int64(nw)
			if ew != nil {
				return n, ew
			}
		}

		if er != nil {
			if errors.Is(er, io.EOF) {
				return n, nil
			}

			return n, er
		}
	}
}

// WriteTo writes all unread [DataInterval] items from the [Timeframe] to w,
// advancing the read cursor and implementing [gio.WriterTo].
func (t *Timeframe[T]) WriteTo(w gio.Writer[DataInterval[T]]) (n int64, err error) {
	if t == nil || w == nil || t.readIdx >= int64(len(t.buffer)) {
		return 0, nil
	}

	remaining := t.buffer[t.readIdx:]
	m, err := w.Write(remaining)
	if m < 0 || m > len(remaining) {
		return n, gio.ErrInvalidWrite
	}

	t.readIdx += int64(m)
	n = int64(m)

	if err != nil {
		return n, err
	}

	if m != len(remaining) {
		return n, gio.ErrShortWrite
	}

	return n, nil
}

// Reset clears all organized intervals and resets the read offset to 0 on the
// [TimeframeStream], retaining the underlying storage capacity for future writes.
func (t *TimeframeStream[T]) Reset() {
	if t == nil {
		return
	}

	clear(t.buffer)
	t.buffer = t.buffer[:0]
	t.readIdx = 0
	t.seeded = false
}

// Write inserts each [DataInterval] in p into the [TimeframeStream], resolving
// overlaps eagerly on each item and implementing [gio.Writer].
//
// If any element in p carries an invalid [Interval], Write stops at that
// element and returns the number of elements written so far along with
// [ErrInvalidInterval].
func (t *TimeframeStream[T]) Write(p []DataInterval[T]) (n int, err error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	t.init()

	for i := range p {
		if !t.Add(p[i].Interval, p[i].Data) {
			return i, ErrInvalidInterval
		}
	}

	return len(p), nil
}

// WriteItem inserts a single [DataInterval] into the [TimeframeStream],
// resolving overlaps eagerly and implementing [gio.ItemWriter]. It returns
// [ErrInvalidInterval] if item's interval is invalid.
func (t *TimeframeStream[T]) WriteItem(item DataInterval[T]) error {
	if t == nil {
		return ErrNilReceiver
	}

	if !t.Add(item.Interval, item.Data) {
		return ErrInvalidInterval
	}

	return nil
}

// Read copies up to len(p) unread organized [DataInterval] items from the
// [TimeframeStream] into p, advancing the internal read cursor and implementing
// [gio.Reader].
//
// The underlying buffer is retained so subsequent writes resolve conflicts
// against the full timeline and [TimeframeStream.All], [TimeframeStream.Slice],
// and [TimeframeStream.Seek] continue to access all organized intervals.
func (t *TimeframeStream[T]) Read(p []DataInterval[T]) (n int, err error) {
	if t == nil {
		return 0, io.EOF
	}

	if len(p) == 0 {
		return 0, nil
	}

	if t.readIdx >= int64(len(t.buffer)) {
		return 0, io.EOF
	}

	n = copy(p, t.buffer[t.readIdx:])
	t.readIdx += int64(n)

	return n, nil
}

// ReadAt copies up to len(p) organized [DataInterval] items starting at offset
// off in the [TimeframeStream] buffer without modifying the read cursor,
// implementing [gio.ReaderAt].
func (t *TimeframeStream[T]) ReadAt(p []DataInterval[T], off int64) (n int, err error) {
	if off < 0 {
		return 0, ErrNegativeOffset
	}

	if t == nil || off >= int64(len(t.buffer)) {
		return 0, io.EOF
	}

	n = copy(p, t.buffer[off:])
	if n < len(p) {
		err = io.EOF
	}

	return n, err
}

// ReadItem reads and returns the next unread organized [DataInterval] from the
// [TimeframeStream], advancing the read cursor by 1 and implementing
// [gio.ItemReader].
func (t *TimeframeStream[T]) ReadItem() (DataInterval[T], error) {
	if t == nil || t.readIdx >= int64(len(t.buffer)) {
		var zero DataInterval[T]
		return zero, io.EOF
	}

	item := t.buffer[t.readIdx]
	t.readIdx++

	return item, nil
}

// UnreadItem steps the read cursor back by one position, implementing
// [gio.ItemScanner]. It returns [ErrUnreadAtStart] if the cursor is already at
// index 0.
func (t *TimeframeStream[T]) UnreadItem() error {
	if t == nil || t.readIdx <= 0 {
		return ErrUnreadAtStart
	}

	t.readIdx--

	return nil
}

// Seek sets the offset for the next Read or WriteTo operation on the
// [TimeframeStream], implementing [gio.Seeker].
func (t *TimeframeStream[T]) Seek(offset int64, whence int) (int64, error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	var abs int64

	switch whence {
	case gio.SeekStart:
		abs = offset
	case gio.SeekCurrent:
		abs = t.readIdx + offset
	case gio.SeekEnd:
		abs = int64(len(t.buffer)) + offset
	default:
		return 0, ErrInvalidWhence
	}

	if abs < 0 {
		return 0, ErrNegativePosition
	}

	t.readIdx = abs

	return abs, nil
}

// ReadFrom reads [DataInterval] items from r until EOF or error and inserts
// them into the [TimeframeStream], resolving overlaps eagerly and implementing
// [gio.ReaderFrom].
func (t *TimeframeStream[T]) ReadFrom(r gio.Reader[DataInterval[T]]) (n int64, err error) {
	if t == nil {
		return 0, ErrNilReceiver
	}

	if r == nil {
		return 0, nil
	}

	buf := make([]DataInterval[T], minAlloc)

	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := t.Write(buf[:nr])
			n += int64(nw)
			if ew != nil {
				return n, ew
			}
		}

		if er != nil {
			if errors.Is(er, io.EOF) {
				return n, nil
			}

			return n, er
		}
	}
}

// WriteTo writes all unread organized [DataInterval] items from the
// [TimeframeStream] to w, advancing the read cursor and implementing
// [gio.WriterTo].
func (t *TimeframeStream[T]) WriteTo(w gio.Writer[DataInterval[T]]) (n int64, err error) {
	if t == nil || w == nil || t.readIdx >= int64(len(t.buffer)) {
		return 0, nil
	}

	remaining := t.buffer[t.readIdx:]
	m, err := w.Write(remaining)
	if m < 0 || m > len(remaining) {
		return n, gio.ErrInvalidWrite
	}

	t.readIdx += int64(m)
	n = int64(m)

	if err != nil {
		return n, err
	}

	if m != len(remaining) {
		return n, gio.ErrShortWrite
	}

	return n, nil
}
