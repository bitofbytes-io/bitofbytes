# Timeline

The project's history in the detail page aside, newest first. It is the only place the latest update appears on a detail page.

- `ol.nc-timeline` with a 2px `line-strong` rail. Each `nc-timeline__item` has a `meta` date and one sentence in `ink` 16px.
- The newest item adds `nc-timeline__item--latest` (filled `accent` dot) and a `<b>latest</b>` tag after its date. Older items have hollow `line-control` dots.
- The last item is always the first commit ("First commit.", from `FirstCommitDate`).
- Today each project has one note, so the timeline has two items. When the data grows a list of dated updates, show them all here, newest first; don't add a separate "latest update" box elsewhere on the page.
