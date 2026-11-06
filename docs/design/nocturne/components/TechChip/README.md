# TechChip

Small outlined label for one technology in a project's stack, plus the Badge that marks recent work.

- **Chip** (`nc-chip` inside a `ul.nc-chips` with an `aria-label` like "Noted technology"): `mono` 12px `ink-muted`, `line-strong` border, `radius-sm`, 4 × 10px padding, 8px gaps. Text exactly as `Tech` lists it.
- Project rows show the first four; detail pages show all of them (13px, 6 × 12px padding).
- **Chip links.** On the Projects list and the project page each chip is an `<a class="nc-chip">` inside its `li`, linking to `/projects?tech=<name>` (the list keeps the current `sort`). The hover color is `accent`.
- **Filter pill** (`nc-filter`): shown above the Projects list while `?tech=` is set. A `surface` pill reading "Technology" plus the name in mono, then a "Clear" `accent` link (44px tall) back to the unfiltered list. The eyebrow becomes "Projects · 2 of 8".
- **Empty state** (`nc-empty`): when no project uses the technology, a single `ink-soft` line, "No projects use X yet.", with a link back to all projects. Not an error page.
- **Badge** (`nc-badge`): `surface-accent` pill with an `accent` dot and label. The only badge is "Updated recently", beside a project name in a Project row. Don't invent other statuses.
