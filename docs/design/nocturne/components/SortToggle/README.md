# SortToggle

Two-option segmented control that orders the Projects list.

- `role="group"` with an `aria-label`, holding two `<button aria-pressed>` elements. The pressed one takes the `accent` fill and `on-accent` label.
- Options: "Recently updated" (by `LastUpdate`, the default) and "Newest first" (by `FirstCommitDate`). Without JavaScript, render each option as a link to `?sort=updated` / `?sort=newest` and mark the current one.
- Sits at the right end of the page header row on desktop; full width under the intro on phone.
