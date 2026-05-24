FROM node:24.15.0-alpine3.23 AS tailwind-builder
ARG TAILWIND_VERSION=4.3.0
WORKDIR /app
RUN npm init -y \
    && npm install tailwindcss@${TAILWIND_VERSION} @tailwindcss/cli@${TAILWIND_VERSION} @tailwindcss/typography
COPY ./templates ./templates
COPY ./tailwind/styles.css src/styles.css
RUN npx @tailwindcss/cli -i src/styles.css -o /styles.css --minify

FROM golang:1.26.3-alpine3.23 AS builder
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
COPY ./static ./static
COPY --from=builder /app/bob ./bob
COPY --from=tailwind-builder /styles.css ./static/styles.css
RUN addgroup -S bob \
    && adduser -S -G bob bob \
    && chown -R bob:bob /bob

LABEL org.opencontainers.image.source="${SOURCE_URL}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="bitofbytes" \
      org.opencontainers.image.description="BitOfBytes web application"

ENV CSRF_KEY_FILE=/run/secrets/csrf_key
ENV SERVER_ADDRESS=:3000
ENV CSRF_SECURE=true
ENV LOG_LEVEL=${LOG_LEVEL}

USER bob

EXPOSE 3000
CMD ["./bob"]
