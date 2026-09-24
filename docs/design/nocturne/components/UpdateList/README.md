# UpdateList

The card beside the Home hero that answers "what has Daniel been working on?"

- An `nc-card` with class `nc-updates`: an uppercase `nc-card__title` "Latest updates" and an `nc-live` count ("6 this month": projects whose `LastUpdate` is within 30 days; omit it when zero).
- Four rows, newest `LastUpdate` first. Each row is one link (`nc-update`) to the project's detail page: name (Plex Sans 600 17px), short date (`meta`, "Sep 6"), and a one-sentence note in `ink-muted` 15px.
- Notes are the project's `Notes` with the "Most recent work…" lead-in removed and trimmed to one short sentence.
- 440px wide on desktop, full width on phone (three rows there).
