FROM golang:latest AS builder

WORKDIR /go/src/github.com/zalgonoise/x/modupdate

COPY ./ ./

RUN go mod download
RUN mkdir /build \
    && go build -o /build/modupdate ./cmd/modupdate \
    && chmod +x /build/modupdate

# TODO: get a nonroot or alpine image going instead of debiand
FROM debian:latest

WORKDIR /app/data/go

COPY --from=builder /build/modupdate /app/modupdate

RUN ln -s -f /app/data/go /go

RUN touch /app/data/sqlite.db

COPY ./ /go/src/github.com/zalgonoise/x/modupdate

WORKDIR /

CMD ["/app/modupdate"]