# Button

Pill-shaped link or button for the one or two actions a section offers.

- **Primary** (`nc-btn nc-btn--primary`): `accent` fill, `on-accent` label, weight 600. One per view: "See the projects" on Home, "Open live site" on a project page.
- **Secondary** (`nc-btn nc-btn--secondary`): transparent with a `line-control` border and `ink` label: "Resume (PDF)", "View on GitHub".
- 52px tall (`control-height`), `radius-pill`, 26px side padding, `body` size. A trailing 18px icon is allowed: arrow-right for in-site navigation, external-link for leaving the site.
- Group buttons in `nc-actions` (16px gap). Below 640px they stack full width.
- Use `<a>` when it navigates, `<button>` when it acts. Labels are sentence case verbs; never "Click here".
