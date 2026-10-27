package poller

import "github.com/zalgonoise/x/gacha/internal/errs"

// Sentinel error kinds and entities used by the poller.
const (
	domain = "gacha/poller"

	ErrFailedTo  = errs.Kind("failed to")
	ErrRecovered = errs.Kind("recovered")
	ErrExceeded  = errs.Kind("exceeded")

	ErrClaimEntries   = errs.Entity("claim entries")
	ErrSettleEntries  = errs.Entity("settle entries")
	ErrExtendLeases   = errs.Entity("extend leases")
	ErrUnmarshalEntry = errs.Entity("unmarshal entry")
	ErrHandlerPanic   = errs.Entity("handler panic")
	ErrAttempts       = errs.Entity("max attempts")
)

// Sentinel errors reported by the poller.
var (
	// ErrClaim wraps the error of a failed Source.Claim.
	ErrClaim = errs.New(domain, ErrFailedTo, ErrClaimEntries)
	// ErrSettle wraps the error of a failed Source.Settle. The entries stay leased until their leases expire, and
	// are then claimed again.
	ErrSettle = errs.New(domain, ErrFailedTo, ErrSettleEntries)
	// ErrExtend wraps the error of a failed gacha.Extender.Extend heartbeat.
	ErrExtend = errs.New(domain, ErrFailedTo, ErrExtendLeases)
	// ErrUnmarshal wraps the error of a failed Unmarshal. Such entries are moved to the dead-letter table.
	ErrUnmarshal = errs.New(domain, ErrFailedTo, ErrUnmarshalEntry)
	// ErrPanic wraps the value of a recovered handler panic. Panics are retried like errors.
	ErrPanic = errs.New(domain, ErrRecovered, ErrHandlerPanic)
	// ErrMaxAttempts is recorded for entries claimed more times than the maximum, which are moved to the dead-letter
	// table without running the handler. This happens when handlers crash the process before their entry is settled.
	ErrMaxAttempts = errs.New(domain, ErrExceeded, ErrAttempts)
)
