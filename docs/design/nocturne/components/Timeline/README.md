# Timeline

The project's history in the detail page aside, newest first. It is the only place the latest update appears on a detail page.

- `ol.nc-timeline` with a 2px `line-strong` rail. Each `nc-timeline__item` has a `meta` date and one sentence in `ink` 16px.
- The newest item adds `nc-timeline__item--latest` (filled `accent` dot) and a `<b>latest</b>` tag after its date. Older items have hollow `line-control` dots.
- On a project page, show up to three dated update notes, newest first. The latest note comes from `LastUpdate` and `Notes`; the two older notes come from `PreviousUpdates`. Do not show a generic first-commit item or add a separate "latest update" box elsewhere on the page.

## Updates page

`/updates` lists every update from `models.AllUpdates()` in one `ol.nc-timeline.nc-timeline--page` (720px wide, 32px between items, 24px on phone). Each item wraps its content in an `a.nc-timeline__entry` to the project page, so the whole item is the touch target: the date (with "latest" on the first item only), the project name (`nc-timeline__name`, Plex Sans 600 17px `accent`, underlined on hover like the ledger), then the note.

`?project=<slug>` narrows the page to one project. A `nav.nc-filter` above the timeline holds a chip row (`ul.nc-chips`): "All projects" plus one `a.nc-chip` per project, and the active one carries `aria-current="true"` (`accent` fill, `on-accent` text). The eyebrow becomes "Changelog · 3 of 22 updates", the title "Carma *updates*", and the intro links that project's feed. An unknown slug is ignored and shows the full list.

The project page adds one `p.nc-timeline__more` under its Timeline: "All Carma updates" (to `/updates?project=<slug>`) and "Atom feed" (to `/projects/<slug>/feed.xml`), `ink-muted` 15px with `nc-prose-link` links.
