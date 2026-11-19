FROM golang:1.27.4-alpine3.23@sha256:f70850ef6ce341e8b6427f677049145d63c0b7ce002d4d7e3d464828b1b96dcd AS builder
ARG VERSION=dev
ARG REVISION=unknown
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X main.version=${VERSION} -X main.revision=${REVISION}" -v -o ./bob ./cmd/bob/

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
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
    CMD wget -qO- "http://localhost:${SERVER_ADDRESS##*:}/healthz" >/dev/null || exit 1
CMD ["./bob"]
