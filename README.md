# BitOfBytes

BitOfBytes is the Go web application behind the BitOfBytes portfolio. It renders project pages on the server, serves embedded templates and static assets, and has no database or third-party API dependency.

## Requirements

- Docker 24+

## Build the image

The multi-stage Dockerfile builds the Go binary and copies the static assets:

```bash
docker build -f Docker/Dockerfile -t bitofbytes:local .
```

## Configure the application

Create the ignored `.env` file:

```dotenv
SERVER_ADDRESS=:3000
CSRF_SECURE=false
LOG_LEVEL=info
LOG_FORMAT=text
```

Do not commit this file.

| Setting | Required | Purpose |
| --- | --- | --- |
| `SERVER_ADDRESS` | Yes | Listen address inside the container; use `:3000` |
| `CSRF_SECURE` | Yes | Set `false` for local HTTP and `true` behind production HTTPS |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, or `error` |
| `LOG_FORMAT` | No | `text` or `json`; defaults to `text` |

## Run with Docker

```bash
docker run --rm --name bitofbytes \
  --env-file .env \
  -p 3000:3000 \
  bitofbytes:local
```

Open <http://localhost:3000>. The health endpoint is <http://localhost:3000/healthz>.

For production, terminate TLS at a reverse proxy, and set `CSRF_SECURE=true`.

## Development

Copy the template configuration:

```bash
cp .env.template .env
# Set SERVER_ADDRESS=:3000 and CSRF_SECURE=false.
go run ./cmd/bob
```

For live reload, install `air`, then run:

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
