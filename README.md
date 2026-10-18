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

Open <http://localhost:3000>. Two probe endpoints follow the shared health contract:

- `/health` returns `200 ok` while the process is up. `/healthz` is the same handler and stays for the image's `HEALTHCHECK`.
- `/ready` reports whether dependencies are up. BitOfBytes has no database or mount, so it always returns `200 {"status":"ok"}`.

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

### SEO metadata

Each page renders its own title, description, canonical URL, Open Graph and Twitter tags, and a JSON-LD block from `views.Meta`, which the page data in `controllers/portfolio.go` embeds. Project pages take their metadata from `models.Project`. `/robots.txt` is `static/robots.txt`, and `/sitemap.xml` is generated from `models.Projects()` with each project's last update as `<lastmod>`. The canonical host is hardcoded to `https://bitofbytes.io` (`views.SiteURL`); no configuration is needed.

## Analytics events

Umami page views load through `static/umami-loader.js` on `bitofbytes.io` only. Outbound intent is recorded with Umami's `data-umami-event` attributes on the links themselves, so no inline JavaScript is needed and the CSP is unchanged. There is no new configuration; the attributes are inert wherever the Umami script does not load.

| Event | Link | Properties |
| --- | --- | --- |
| `live-click` | Open live site on a project page | `project` (slug) |
| `repo-click` | View on GitHub on a project page | `project` (slug) |
| `screenshot-click` | Screenshot on a project page | `project` (slug) |
| `resume-click` | Resume PDF links | `location` (`header`, `menu`, `home-hero`) |
| `email-click` | `mailto:` link in the footer | `location` (`footer`) |

Add the same attributes to new outbound links to track them.

## License

BitOfBytes is available under the [MIT License](LICENSE).
