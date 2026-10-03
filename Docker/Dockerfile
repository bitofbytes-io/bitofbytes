FROM golang:1.26.6-alpine3.23 AS builder
ARG VERSION=dev
ARG REVISION=unknown
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X main.version=${VERSION} -X main.revision=${REVISION}" -v -o ./bob ./cmd/bob/

FROM alpine:3.23.4
ARG LOG_LEVEL=warn
ARG VERSION=dev
ARG REVISION=unknown
ARG SOURCE_URL=https://github.com/bitofbytes-io/bitofbytes
WORKDIR /bob
# The binary and static files stay owned by root, so the bob user can read
# them but not change them.
COPY ./static ./static
COPY --from=builder /app/bob ./bob
RUN addgroup -S bob \
    && adduser -S -G bob bob

LABEL org.opencontainers.image.source="${SOURCE_URL}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="bitofbytes" \
      org.opencontainers.image.description="BitOfBytes web application"

ENV SERVER_ADDRESS=:3000
ENV CSRF_SECURE=true
ENV LOG_LEVEL=${LOG_LEVEL}

USER bob

EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:3000/healthz >/dev/null || exit 1
CMD ["./bob"]
