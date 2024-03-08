FROM golang:alpine AS builder

WORKDIR /go/scr/github.com/zalgonoise/x/authz

COPY ./ ./

RUN go mod download
RUN mkdir /build \
    && go build -ldflags="-s -w" -o /build/authz ./cmd/authz \
    && chmod +x /build/authz

FROM alpine:edge

WORKDIR /app

COPY --from=builder /build/authz /app/authz

CMD ["/app/authz"]