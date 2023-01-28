FROM golang:alpine AS builder

WORKDIR /go/src/github.com/zalgonoise/x/secr

COPY ./ ./

# this app's sqlite requires gcc 
RUN apk add build-base

RUN go mod download
RUN mkdir /build \
    && go build -o /build/secr . \
    && chmod +x /build/secr


FROM alpine:edge

RUN mkdir -p /secr/server

COPY --from=builder /build/secr /secr

CMD ["/secr"]