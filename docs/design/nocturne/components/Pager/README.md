# Pager

Previous and next project links at the foot of a project page, after the Screenshots gallery.

- `nav aria-label="More projects"` holding two links in a two-column grid, separated from the page above by a `line` rule. Previous sits left, Next right and right-aligned.
- Each link is a `mono` 14px label in `ink-subtle` ("Previous project", "Next project") over the project name in `display` (28px, 22px below 640px) in `ink`. Hover turns the name `accent`. The link is at least `control-height-sm` tall.
- Neighbours follow the catalogue order in `models.Projects()` and wrap: the first project's Previous is the last project, and the last project's Next is the first.
- Hidden when there is no other project to link to.
