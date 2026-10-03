# Agent Guidance

- `static/styles.css` is hand-written and served as is; there is no CSS build step.
- Preserve CSRF protection: state-changing routes must use POST, PUT, PATCH, or DELETE (never GET) so the `middleware.CSRF` cross-origin check covers them; forms need no token field.
- Keep portfolio content in `models/project.go` and follow the existing embedded-template pattern for new pages.
- After configuring the ignored `.env`, use `go run ./cmd/bob` for a one-shot local server or `make local` for live reload with `air`.
- Run `go test ./...` for changes and verify rendered routes when templates, middleware, or static assets change.
- The UI follows the Nocturne design system in `docs/design/nocturne/`: read its `README.md` before changing templates or styles. For visual changes, update `tokens.json` or `components/bundle.css` there first, keep `tokens.css` in sync with token edits, and mirror the result into `static/styles.css`.
- Site copy never uses em dashes.
- Project screenshots under `static/projects/<slug>/` are WebP (quality 90) and the `Path` in `models/project.go` points at the `.webp`. Capture new desktop screenshots at a 1280 by 900 viewport, not full page (some older captures predate this and keep their size until they are refreshed); device captures keep their native size. Encode with Chromium's canvas encoder (Playwright) or `cwebp -q 90`; do not add PNG screenshots.
