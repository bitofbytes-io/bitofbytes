# Agent Guidance

- Edit `tailwind/styles.css`, not generated `static/styles.css`; rebuild it before validating production output.
- Preserve CSRF protection for state-changing routes and include the CSRF template field in new forms.
- Keep portfolio content in `models/project.go` and follow the existing embedded-template pattern for new pages.
- After configuring the ignored `.env`, use `go run ./cmd/bob` for a one-shot local server or `make local` for live reload with Tailwind watch and `air`.
- Build production CSS with `make tail-prod`; the Dockerfile also builds Tailwind during image creation.
- Run `go test ./...` for changes and verify rendered routes when templates, middleware, or static assets change.
- The UI follows the Nocturne design system in `docs/design/nocturne/`: read its `README.md` before changing templates or styles. Keep `tailwind/styles.css` in sync with `docs/design/nocturne/tokens.css` and `components/bundle.css`, and `static/nocturne.js` in sync with `components/bundle.js`.
- Site copy never uses em dashes.
