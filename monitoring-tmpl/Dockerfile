FROM golang:alpine3.17 AS builder

WORKDIR /go/src/github.com/zalgonoise/x/monitoring-tmpl

COPY ./ ./

RUN go mod download
RUN mkdir /build \
    && go build -o /build/app ./cmd/exemplars-ex \
    && chmod +x /build/app


FROM alpine:edge

COPY --from=builder /build/app /.
CMD [ "/app" ]