package log

import (
	"log/slog"
)

type SpanContextHandlerOption interface {
	apply(handler *SpanContextHandler)
}

type optWithSpanID struct{}

func (optWithSpanID) apply(handler *SpanContextHandler) {
	handler.withSpanID = true
}

func WithSpanID() SpanContextHandlerOption {
	return optWithSpanID{}
}

type optWithHandler struct {
	handler slog.Handler
}

var zeroHandler slog.Handler

func (o *optWithHandler) apply(handler *SpanContextHandler) {
	if o.handler == nil || o.handler == zeroHandler {
		o.handler = defaultHandler()
	}

	handler.handler = o.handler
}

func WithHandler(handler slog.Handler) SpanContextHandlerOption {
	return &optWithHandler{
		handler: handler,
	}
}
