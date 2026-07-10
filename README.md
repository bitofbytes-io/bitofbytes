# BitOfBytes

BitOfBytes is the Go web application behind the BitOfBytes portfolio. It renders project pages on the server, serves embedded templates and static assets, and has no database or third-party API dependency.

## Requirements

- Docker 24+
- OpenSSL or another secure random generator for the CSRF key

## Build the image

The multi-stage Dockerfile builds the Go binary and Tailwind CSS:

```bash
docker build -f Docker/Dockerfile -t bitofbytes:local .
```

## Configure the application

Generate a Base64-encoded 32-byte CSRF key:

```bash
openssl rand -base64 32
```

Create the ignored `.env` file and paste the generated value:

```dotenv
SERVER_ADDRESS=:3000
CSRF_KEY=replace-with-generated-value
CSRF_SECURE=false
LOG_LEVEL=info
LOG_FORMAT=text
```

Do not commit this file.

| Setting | Required | Purpose |
| --- | --- | --- |
| `SERVER_ADDRESS` | Yes | Listen address inside the container; use `:3000` |
| `CSRF_KEY` | One of | Base64 value that decodes to exactly 32 bytes |
| `CSRF_SECURE` | Yes | Set `false` for local HTTP and `true` behind production HTTPS |
| `CSRF_KEY_FILE` | One of | Read the CSRF key from a mounted file instead of `CSRF_KEY` |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | No | `text` or `json`; defaults to `text` |

The container defaults `CSRF_KEY_FILE` to `/run/secrets/csrf_key`, so a secret mount can be used instead of an environment value.

## Run with Docker

```bash
docker run --rm --name bitofbytes \
  --env-file .env \
  -p 3000:3000 \
  bitofbytes:local
```

Open <http://localhost:3000>. The health endpoint is <http://localhost:3000/healthz>.

For production, terminate TLS at a reverse proxy, set `CSRF_SECURE=true`, and provide the CSRF key through your deployment platform's secret manager.

## Development

Copy the template configuration and generate a local key:

```bash
cp .env.template .env
# Set SERVER_ADDRESS=:3000, CSRF_SECURE=false, and CSRF_KEY to the generated value.
go run ./cmd/bob
```

For live reload, install `air` and the Tailwind CSS CLI, then run:

```bash
make local
```

Run the test suite with:

```bash
go test ./...
```

Portfolio content is maintained in `models/project.go`; templates and static assets live under `templates/` and `static/`.

## License

BitOfBytes is available under the [MIT License](LICENSE).
