# Field Notebook Site Plan

## Design thesis

Field Notebook should feel like a crafted engineering notebook: warm, technical, personal, and readable. The site should keep a little code-and-signal character, but the main experience should no longer depend on pixel fonts, neon borders, or arcade framing.

## Visual system

### Fonts

- Display: `Newsreader`
  - Use for major identity moments: home hero name, project detail title, and large page headings.
  - Weight target: 500 to 650.
  - Keep line-height tight, around `0.9` to `1.0`, but avoid oversized mobile text.
- Body and UI: `Manrope`
  - Use for paragraphs, navigation, links, cards, buttons, captions, and general UI.
  - Weight target: 400 for body, 700 to 800 for commands and nav.
- Technical labels: `IBM Plex Mono`
  - Use only for eyebrows, metadata labels, dates, tech tags, and short status text.
  - Avoid using it for long paragraphs.

### Color tokens

```css
:root {
  --bob-bg: #1a1714;
  --bob-surface: #211c18;
  --bob-surface-2: #241f1a;
  --bob-ink: #f7f0e5;
  --bob-muted: #b9ab99;
  --bob-soft: #8e8173;
  --bob-line: #3b3129;
  --bob-gold: #f7bf4f;
  --bob-teal: #5fd0be;
  --bob-plum: #ad74ff;
}
```

Usage:

- `--bob-bg`: body background.
- `--bob-surface`: primary panels and section blocks.
- `--bob-surface-2`: image frames, nested metadata rows, screenshot captions.
- `--bob-ink`: main text and high-priority links.
- `--bob-muted`: body copy and secondary text.
- `--bob-soft`: tertiary labels.
- `--bob-line`: borders and dividers.
- `--bob-gold`: primary accent and call-to-action.
- `--bob-teal`: technical/status accent.
- `--bob-plum`: occasional hover/focus accent only.

### Texture and shape

- Use a subtle notebook grid, not a pixel game grid:
  - 48px spacing.
  - Gold and teal lines at very low opacity.
- Border radius stays at `8px`.
- Borders should be thin and quiet: `1px solid var(--bob-line)`.
- Avoid heavy corner brackets, scanlines, hard neon outlines, and text shadows.
- Use soft contrast through panels, dividers, and image framing rather than glowing effects.

### Type scale

- Hero/page H1: `clamp(4rem, 9vw, 8.4rem)` desktop, capped lower on mobile.
- Project index H1: `clamp(3.2rem, 7vw, 6rem)`.
- Card/project titles: `clamp(2rem, 4vw, 3.8rem)`.
- Section H2: `clamp(1.6rem, 3vw, 2.4rem)`.
- Body: `1rem` to `1.12rem`, line-height `1.7` to `1.8`.
- Metadata: `0.76rem` to `0.85rem`, uppercase mono.

## Shared components

### Header and navigation

- Keep the current simple nav structure.
- Replace the pixel brand treatment with a quieter notebook lockup:
  - `BoB` mark: 42px square, 1px gold border, no pixel font.
  - Brand text: Manrope 800.
  - Nav links: Manrope 800, muted by default, gold or teal on active/hover.
- Header background should be translucent dark paper, not black monitor glass.

### Panels

Replace `.pixel-panel` with a more general notebook panel class, likely while keeping the old class name temporarily for migration:

- `background: var(--bob-surface)`
- `border: 1px solid var(--bob-line)`
- `border-radius: 8px`
- no decorative corner pseudo-elements
- no text-shadow

### Buttons

Replace `.pixel-button` styling with notebook actions:

- Primary: gold background, dark text, no uppercase requirement unless label is technical.
- Secondary: transparent/dark surface with quiet border.
- Focus state: visible teal or gold outline.
- Keep minimum target size at 44px.

### Tags and metadata

- Tech tags become small notebook labels:
  - border: `1px solid rgb(95 208 190 / 0.28)`
  - text: teal
  - background: low-opacity teal
  - mono, uppercase, compact
- Metadata rows use dark-surface rows with a mono label and readable value.

## Page application

### Home

Goal: make Field Notebook the canonical first impression.

Structure:

- Hero uses the Field Notebook layout:
  - left: large `Daniel Waters`, short role line, paragraph, primary actions.
  - right: profile photo plus two note rows.
- Use `daniel_train.jpg` by default for the human/professional feel.
- Keep the pixel avatar only if we want a small optional easter egg or secondary asset later.
- Contact links should live in compact action rows, not large arcade cards.

Content changes:

- Replace "Senior Software Engineer @ Cisco" as the main role line if desired with:
  - `Notes from software, video, and useful side projects.`
- Keep Cisco and Raleigh as metadata notes rather than forcing everything into the H1 area.

### Projects index

Goal: make the project list feel like notebook entries or field notes.

Structure:

- Page heading becomes a notebook intro panel.
- Each `.project-listing` becomes an entry:
  - left: date/update eyebrow, title, summary.
  - right: tech tags plus a single details action.
- Use a small left accent rule or top divider, not neon boxes.
- Consider changing "Last update" to "Updated" to feel less system-y.

Visual treatment:

- Project names in `Newsreader`.
- Summaries in Manrope with comfortable line height.
- Tech lists in IBM Plex Mono.

### Project detail

Goal: make each project page read like a dossier page inside the same notebook.

Structure:

- Hero:
  - "Project dossier" eyebrow can stay, but style it as a notebook label.
  - Project title in `Newsreader`.
  - Tagline in teal or muted ink, not green pixel type.
  - GitHub/live links styled as notebook actions.
- Main content:
  - Overview, Stack notes, and Highlights become sections within a single readable article panel.
  - Sidebar metadata becomes a "Field notes" panel.
- Screenshots:
  - Screenshot frames should look like paper/photo inserts:
    - dark surface frame
    - quiet border
    - caption in Manrope or mono label plus title
  - Keep screenshots large and inspection-friendly.

### Mobile behavior

- Collapse all two-column layouts to one column below `840px`.
- Keep nav readable with the existing details menu if needed, but restyle it to match notebook controls.
- H1 caps should prevent long project names from wrapping awkwardly.
- Contact/action buttons should become full-width only where it improves tap ergonomics.

## Implementation order

1. Update global font import in `templates/base.gohtml`.
2. Replace root CSS variables in `tailwind/styles.css` with Field Notebook tokens.
3. Restyle shared shell, header, nav, panel, button, tag, metadata, and screenshot primitives.
4. Update homepage template classes/content to match the Field Notebook structure.
5. Update projects index styles using existing template structure first.
6. Update project detail styles and, only if needed, lightly adjust markup for "Field notes" and screenshot captions.
7. Run `make tail-prod` to rebuild `static/styles.css`.
8. Run `go test ./...`.
9. Verify `/`, `/projects`, and at least one `/projects/{slug}` page at desktop and mobile widths.

## Implementation notes

- Keep the Go data model unchanged.
- Avoid adding JavaScript for this redesign.
- Do not introduce generated image assets unless a later pass needs a custom texture or portrait treatment.
- Keep the old design board as reference, but the production CSS should only carry one design system.
- After the theme is applied, recapture BitOfBytes project screenshots because the current project screenshots show the Pixel Workbench design.

## Acceptance checklist

- No pixel display font remains in normal UI or paragraph content.
- Body copy is readable without zooming on mobile.
- Home, projects index, and project detail pages clearly belong to the same visual system.
- Project screenshots remain easy to inspect.
- Focus states are visible on all links and buttons.
- The design feels personal and technical, not like a generic SaaS template.
