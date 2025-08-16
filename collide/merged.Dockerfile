FROM gcr.io/distroless/static-debian11:nonroot
COPY out/app /app
COPY out/app-fe /app-fe

COPY internal/tracks/tracks.yaml /tracks.yaml

ENTRYPOINT ["/app", "serve", "&", "/app-fe", "serve"]