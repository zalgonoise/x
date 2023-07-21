FROM golang:alpine3.17 AS builder


WORKDIR /go/src/github.com/zalgonoise/x/audio

COPY ./ ./

RUN go mod download 
RUN mkdir /build \
    && go build -o /build/audio ./cmd/audio \
    && chmod +x /build/audio


FROM alpine:edge

COPY --from=builder /build/audio /.
CMD [ "/audio" ]