# Agent Guidance

- Edit `tailwind/styles.css`, not generated `static/styles.css`; rebuild it before validating production output.
- Preserve CSRF protection for state-changing routes and include the CSRF template field in new forms.
- Keep portfolio content in `models/project.go` and follow the existing embedded-template pattern for new pages.
- Run `go test ./...` for changes and verify rendered routes when templates, middleware, or static assets change.
