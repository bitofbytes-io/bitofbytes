# Timeline

The project's history in the detail page aside, newest first. It is the only place the latest update appears on a detail page.

- `ol.nc-timeline` with a 2px `line-strong` rail. Each `nc-timeline__item` has a `meta` date and one sentence in `ink` 16px.
- The newest item adds `nc-timeline__item--latest` (filled `accent` dot) and a `<b>latest</b>` tag after its date. Older items have hollow `line-control` dots.
- Show up to three dated update notes, newest first. The latest note comes from `LastUpdate` and `Notes`; the two older notes come from `PreviousUpdates`. Do not show a generic first-commit item or add a separate "latest update" box elsewhere on the page.
