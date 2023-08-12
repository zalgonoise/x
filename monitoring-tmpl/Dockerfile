FROM golang:alpine3.18 AS builder

WORKDIR /go/src/github.com/zalgonoise/x/monitoring-tmpl

COPY ./ ./

RUN go mod download
RUN mkdir /build \
    && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/app ./cmd/exemplars \
    && chmod +x /build/app


FROM gcr.io/distroless/static

COPY --from=builder /build/app /
CMD [ "/app" ]