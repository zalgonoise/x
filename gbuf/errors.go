package gbuf

import "github.com/zalgonoise/x/gbuf/errs"

const (
	libDomain        = "gbuf"
	bufferDomain     = "gbuf/Buffer"
	readerDomain     = "gbuf/Reader"
	ringBufferDomain = "gbuf/RingBuffer"
	ringFilterDomain = "gbuf/RingFilter"
	peekBufferDomain = "gbuf/PeekBuffer"

	ErrInvalid       = errs.Kind("invalid")
	ErrPreviousOp    = errs.Kind("previous operation")
	ErrNegativeCount = errs.Kind("reader returned negative count")
	ErrTooMuchOf     = errs.Kind("too")
	ErrNegative      = errs.Kind("negative")
	ErrRepeat        = errs.Kind("Repeat")
	ErrIndex         = errs.Kind("index")
	ErrAtBeginning   = errs.Kind("at beginning of")

	ErrWhence           = errs.Entity("whence")
	ErrUnsuccessfulRead = errs.Entity("was not a successful read")
	ErrReadOp           = errs.Entity("from Read")
	ErrLargeSize        = errs.Entity("large")
	ErrCount            = errs.Entity("count")
	ErrWriteCount       = errs.Entity("write count")
	ErrRepeatCount      = errs.Entity("Repeat count")
	ErrOverflow         = errs.Entity("count causes overflow")
	ErrOutOfBounds      = errs.Entity("out of bounds")
	ErrPosition         = errs.Entity("position")
	ErrSlice            = errs.Entity("slice")
	ErrOffset           = errs.Entity("offset")
)

var (
	ErrNegativeRepeatCount = errs.New(libDomain, ErrNegative, ErrRepeatCount)
	ErrCountOverflows      = errs.New(libDomain, ErrRepeat, ErrOverflow)

	ErrBufferUnreadItem     = errs.New(bufferDomain+".UnreadItem", ErrPreviousOp, ErrUnsuccessfulRead)
	ErrRingBufferUnreadItem = errs.New(ringBufferDomain+".UnreadItem", ErrPreviousOp, ErrUnsuccessfulRead)
	ErrRingFilterUnreadItem = errs.New(ringFilterDomain+".UnreadItem", ErrPreviousOp, ErrUnsuccessfulRead)

	ErrReaderInvalidWhence     = errs.New(readerDomain+".Seek", ErrInvalid, ErrWhence)
	ErrRingBufferInvalidWhence = errs.New(ringBufferDomain+".Seek", ErrInvalid, ErrWhence)
	ErrRingFilterInvalidWhence = errs.New(ringFilterDomain+".Seek", ErrInvalid, ErrWhence)

	ErrReaderNegativePosition = errs.New(readerDomain+".Seek", ErrNegative, ErrPosition)

	ErrReaderAtTheBeginning = errs.New(readerDomain+".UnreadItem", ErrAtBeginning, ErrSlice)

	ErrReaderNegativeOffset = errs.New(readerDomain+".ReadAt", ErrNegative, ErrOffset)

	ErrBufferNegativeRead     = errs.New(bufferDomain, ErrNegativeCount, ErrReadOp)
	ErrRingBufferNegativeRead = errs.New(ringBufferDomain, ErrNegativeCount, ErrReadOp)
	ErrRingFilterNegativeRead = errs.New(ringFilterDomain, ErrNegativeCount, ErrReadOp)

	ErrBufferTooLarge = errs.New(bufferDomain, ErrTooMuchOf, ErrLargeSize)

	ErrBufferNegativeCount     = errs.New(bufferDomain+".Grow", ErrNegative, ErrCount)
	ErrReaderNegativeCount     = errs.New(readerDomain+".WriteTo", ErrNegative, ErrCount)
	ErrBufferInvalidWriteCount = errs.New(bufferDomain+".WriteTo", ErrInvalid, ErrWriteCount)

	ErrIndexOutOfBounds           = errs.New(libDomain, ErrIndex, ErrOutOfBounds)
	ErrPeekBufferIndexOutOfBounds = errs.New(peekBufferDomain, ErrIndex, ErrOutOfBounds)
)
