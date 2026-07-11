# Agent Guidance

- Edit `tailwind/styles.css`, not generated `static/styles.css`; rebuild it before validating production output.
- Preserve CSRF protection for state-changing routes and include the CSRF template field in new forms.
- Keep portfolio content in `models/project.go` and follow the existing embedded-template pattern for new pages.
- Use `make local` for live reload; it runs Tailwind watch and `air` together.
- Build production CSS with `make tail-prod`; the Dockerfile also builds Tailwind during image creation.
- Run `go test ./...` for changes and verify rendered routes when templates, middleware, or static assets change.
