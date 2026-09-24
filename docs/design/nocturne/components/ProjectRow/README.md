# ProjectRow

One project in the Projects list: what it is, what changed last, and when.

- Grid of three columns on desktop: 240 × 150px thumbnail (`nc-thumb`), the body, and a 200px side column. 32px vertical padding, `line` rule on top.
- Body: `row-title` name linking to the detail page, followed by a Badge "Updated recently" when `LastUpdate` is within 30 days; the tagline (`ink-soft` 17px); the "Latest ·" line (mono `accent` lead-in, then the note in `ink-subtle`); up to four Tech chips.
- Side: an `nc-meta` list with Updated (full date) and Started (month and year), then a "Details" arrow link with an `aria-label` naming the project.
- Thumbnail: the project's first screenshot, cropped to the top. With no screenshot yet, use `nc-thumb nc-thumb--empty` with the project name in Fraunces italic (the preview shows one).
- Phone: a 96 × 72px thumbnail beside the name; the tagline, latest line, chips and meta run full width underneath.
