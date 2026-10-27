package writer

import "github.com/zalgonoise/x/gacha/internal/errs"

// Sentinel error kinds and entities used by the writers.
const (
	domain = "gacha/writer"

	ErrFailedTo = errs.Kind("failed to")
	ErrUse      = errs.Kind("use of")

	ErrMarshalItem  = errs.Entity("marshal item")
	ErrClosedWriter = errs.Entity("closed writer")
)

// Sentinel errors returned by the writers.
var (
	// ErrMarshal wraps the error of a Marshaler. Nothing is written when an item fails to marshal.
	ErrMarshal = errs.New(domain, ErrFailedTo, ErrMarshalItem)

	// ErrClosed is returned by Batch.Write after Batch.Close was called.
	ErrClosed = errs.New(domain, ErrUse, ErrClosedWriter)
)
