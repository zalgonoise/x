package log

import "context"

func InContext(ctx context.Context, logger Logger) context.Context {
	return nil
}
func From(ctx context.Context) Logger {
	return nil
}
