FROM golang:alpine AS builder

WORKDIR /go/src/github.com/zalgonoise/x/audio

COPY ./ ./

RUN go mod download 
RUN mkdir /build \
    && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/audio ./cmd/audio \
    && chmod +x /build/audio


FROM gcr.io/distroless/static

COPY --from=builder /build/audio /.
CMD [ "/audio" ]