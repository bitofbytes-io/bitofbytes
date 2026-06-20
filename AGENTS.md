# BitOfBytes Project Notes

## Overview
- This repository contains a Go 1.26.4 web application for the BitOfBytes site https://bitofbytes.io
- The HTTP entrypoint is `cmd/bob/bob.go`; it wires up controllers, parses templates from `templates.FS`, and serves static assets from `/static`.
- Controllers live in `controllers/`, services and data types live in `models/`, reusable view helpers live in `views/`, and Go HTML templates are under `templates/`.

## Configuration
- Runtime configuration is loaded from environment variables via `models.LoadEnvConfig()`, which automatically loads a local `.env` file when present.
- Copy `.env.template` to `.env` (the app uses `github.com/joho/godotenv` to load it automatically) and fill in:
  - `SERVER_ADDRESS` (default `:3000` for local development).
  - `CSRF_KEY` (32-byte base64 string recommended) and `CSRF_SECURE` (set `false` for HTTP during local work). In containerized environments you can omit `CSRF_KEY` and instead provide `CSRF_KEY_FILE`, which defaults to `/run/secrets/csrf_key`.
  - `LOG_LEVEL` (debug, info, warn, or error; defaults to info locally and warn in Docker builds) and optionally `LOG_FORMAT` (`text` or `json`).

## Running the server
- Run `go run ./cmd/bob` after preparing the `.env` file to start the web server locally.
- For live-reload development install `air`; use `make run` to start the Go server with `air`, or `make local` to run `air` and the Tailwind watcher (`make tail-watch`) in parallel.
- Tailwind CSS assets are generated from `tailwind/styles.css` into `static/styles.css`. Use `make tail-prod` to build a minified stylesheet for production.

## Content & rendering
- Project portfolio data lives in `models/project.go` and renders through the templates in `templates/projects/`.
- Gorilla CSRF middleware wraps the router; if you add new POST routes make sure to include CSRF tokens in the forms (`csrf.TemplateField`).

## Tests & tooling
- Go unit tests cover controllers, middleware, models, and views. Run `go test ./...` before committing changes and add new tests alongside any new behavior.
- Use `make build` for the local production build pipeline; it runs `make tail-prod`, `make docker-build`, and `make docker-push`.
- Use `make docker-publish` when you only need the Docker build-and-push pair, and `make build-github REGISTRY=<registry> IMAGE_TAG=<tag>` for the CI-equivalent arm64 buildx push. Required registry credentials/secrets must be configured locally or in CI.
- Docker builds use the recipes in the `Makefile` (see `make docker-build` / `make docker-push`) and the multi-stage image defined in `Docker/Dockerfile`.

## Miscellaneous tips
- Static assets (images, PDFs, JS) are stored under `static/` and served at `/static/` via `http.FileServer`.
- Templates use the helper types in `views/` (e.g., `views.Page`, `views.ParseFS`); new templates should follow the existing pattern and be parsed via the views package.
- If you add new template directories, remember that `templates/fs.go` uses a `go:embed` directive, so keep file names glob-friendly.
