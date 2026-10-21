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
| `METRICS_ADDR` | No | Internal Prometheus listener address; defaults to `:9464`, `off` disables it |

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

### Prometheus metrics

`METRICS_ADDR` is trimmed and `off` is case-insensitive. Other values must be a valid IP address or hostname plus a numeric port from 0 to 65535 (IPv6 addresses need brackets). Invalid syntax stops startup; validation does not resolve hostnames. Keep the metrics port private and never publish it through Docker or the public proxy.

`GET /metrics` is served on a separate internal listener at `METRICS_ADDR` (default `:9464`), never on the public `SERVER_ADDRESS` listener, so public proxy routes cannot reach it. The image does not `EXPOSE` the port; scrape it over the internal network. A failure to bind is logged and the site keeps running; `METRICS_ADDR=off` turns the listener off. The code lives in `internal/metrics` and uses its own registry, not the global default.

| Series | Labels | Notes |
| --- | --- | --- |
| `http_requests_total` | `method`, `route`, `status` | Counter for every public request except `/health`, `/healthz`, `/ready` and `/metrics` |
| `http_request_duration_seconds` | `method`, `route`, `status` | Latency histogram with the same labels |
| `go_*`, `process_*` | | Go runtime and process collectors |

`route` is the matched `http.ServeMux` pattern without its method (for example `/projects/{slug}`, `/static/`, or `/{$}` for the home page), and `unmatched` for 404s and 405s, so raw paths and query strings never become label values. `method` is the standard HTTP method, or `OTHER` for any extension method. BitOfBytes has no database, logins, OAuth, uploads, or background jobs, so it exports no DB pool or domain counters.

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

### Updates and Atom feed

Each project's `LastUpdate`/`Notes` and `PreviousUpdates` form a dated changelog. `models.AllUpdates()` flattens them across projects, newest first, and two routes render it:

- `/updates` is the HTML changelog, linked from "See all updates" on the home page.
- `/feed.xml` is the same list as an Atom feed (`application/atom+xml`, cached for an hour). Each entry's ID is `tag:bitofbytes.io,<YYYY-MM-DD>:<slug>`, so keep at most one update per project per day. Every page advertises it with `<link rel="alternate">` and a footer "Subscribe" link.

Adding an update in `models/project.go` updates both on the next deploy; there is no configuration.

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
