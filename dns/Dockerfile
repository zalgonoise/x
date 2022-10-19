FROM golang:alpine AS builder

WORKDIR /go/src/github.com/zalgonoise/x/dns

COPY ./ ./

RUN go mod download
RUN mkdir /build \
    && go build -o /build/dns . \
    && chmod +x /build/dns


FROM alpine:edge

RUN mkdir /conf

COPY --from=builder /build/dns /dns

CMD ["/dns"]