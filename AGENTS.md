# Agent Guidance

- Edit `tailwind/styles.css`, not generated `static/styles.css`; rebuild it before validating production output.
- Preserve CSRF protection: state-changing routes must use POST, PUT, PATCH, or DELETE (never GET) so the `middleware.CSRF` cross-origin check covers them; forms need no token field.
- Keep portfolio content in `models/project.go` and follow the existing embedded-template pattern for new pages.
- After configuring the ignored `.env`, use `go run ./cmd/bob` for a one-shot local server or `make local` for live reload with Tailwind watch and `air`.
- Build production CSS with `make tail-prod`; the Dockerfile also builds Tailwind during image creation.
- Run `go test ./...` for changes and verify rendered routes when templates, middleware, or static assets change.
- The UI follows the Nocturne design system in `docs/design/nocturne/`: read its `README.md` before changing templates or styles. For visual changes, update `tokens.json` or `components/bundle.css` there first, keep `tokens.css` in sync with token edits, mirror the result into `tailwind/styles.css`, and run `make tail-prod`. Keep `static/nocturne.js` in sync with `components/bundle.js`.
- Site copy never uses em dashes.
- Project screenshots under `static/projects/<slug>/` are WebP (quality 90) and the `Path` in `models/project.go` points at the `.webp`. Capture new desktop screenshots at a 1280 by 900 viewport, not full page (some older captures predate this and keep their size until they are refreshed); device captures keep their native size. Encode with Chromium's canvas encoder (Playwright) or `cwebp -q 90`; do not add PNG screenshots.
