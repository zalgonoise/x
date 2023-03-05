FROM golang:alpine3.17 AS builder


WORKDIR /go/src/github.com/zalgonoise/x/audio

COPY ./ ./

RUN go mod download 
RUN mkdir /build \
    && go build -o /build/gsp ./cmd/gsp \
    && chmod +x /build/gsp


FROM alpine:edge

COPY --from=builder /build/gsp /.
CMD [ "/gsp" ]