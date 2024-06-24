package events

type NoOp struct{}

func (NoOp) ReportEvent(Event) {}
func (NoOp) Flush()            {}
