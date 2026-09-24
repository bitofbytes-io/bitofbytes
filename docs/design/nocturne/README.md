# Nocturne

The design system for bitofbytes.io. This folder is the repo copy of the published system at https://claude.ai/artifact/3K9kPLTfxgeoVsRgiZH2ie; the approved page mockups live on the canvas at https://claude.ai/artifact/MyWmR2h4hcnTgwGSB6kbBG.

| Path | What |
| --- | --- |
| `README.md` | This brand book: the rules every page follows. |
| `tokens.json`, `tokens.css` | Colors, type, spacing, radii, shadows and fixed sizes. `tokens.css` is `tokens.json` as CSS custom properties. |
| `components/<Name>/README.md`, `preview.html` | Guidelines and a static preview for each component. Previews expect `tokens.css` and `components/bundle.css` to be loaded. |
| `components/bundle.css` | The component and page-layout styles, mirrored into `tailwind/styles.css`. |
| `components/bundle.js` | The piano keyboard sound, shipped as `static/nocturne.js`. |
| `mockups/*.dc.html` | The approved desktop and phone mockups (Home, Projects, Project detail) as canvas artboards. |
| `screenshots/` | The pages as built, at 1280px. |
| `assets/` | Notes on the example screenshots the published previews use (the images themselves are `static/projects/noted/`). |

When the look changes, update `tokens.json` or `components/bundle.css` here first, mirror it into `tailwind/styles.css`, rebuild with `make tail-prod`, and republish the system so the two copies match.

---

Nocturne is the visual system for bitofbytes.io, the personal site and project portfolio of Daniel Waters. It is a midnight-navy page with one periwinkle accent, a serif display face with italic moments, and a piano keyboard that doubles as the project index. It should feel like a quiet late-night practice session: calm ground, a few lit keys, nothing shouting.

## Content fundamentals

- **Voice.** First person, plain and warm. Daniel talks about his own work: "I build live video and real-time collaboration software at Cisco." Never "we", never marketing superlatives.
- **No em dashes.** Anywhere: headings, copy, taglines, alt text, captions. Use a period, comma or colon instead. "Your permit pal: less yelling, more tracking." A middle dot (`·`) is fine as a separator in eyebrows and meta lines.
- **Casing.** Sentence case for headings and buttons ("See the projects", "Open live site"). The wordmark `bitofbytes` is always lower case. Eyebrows and definition-list labels are set uppercase by CSS; write them in normal case in the markup.
- **Headings with a turn.** Section and page titles may end on an italic accent word: "Side *projects*", "Off the clock, *mostly keys.*", "Daniel / *Waters.*" One italic phrase per heading, never more.
- **Dates.** "Sep 6, 2026" in meta and timelines; "Sep 6" inside the latest-updates list; "since Oct '25" on piano keys. Relative dates ("18 days ago") are not used.
- **Update notes.** One sentence, past tense, what changed for the user: "Improved visit reliability under load and tightened redirect safety." Drop the "Most recent work…" lead-in from the data when displaying it.
- **Placeholders.** Never invent a piece, a game or a stat. In mockups, anything personal Claude does not know stays a visible placeholder in brackets, like `[GAME YOU'RE PLAYING]`; on the live site that content is simply left out until Daniel supplies it.
- **No emoji.** Icons are inline stroke SVG (see Iconography).

## Visual foundations

### Color

- Dark only. Paint every page `bg` with `ink` text. There is no light theme.
- `accent` is the only hue. Spend it on: the primary button, the current nav item, links inside prose, the live dot and "Updated recently" badge, the lit-key edge, one italic word per heading, focus rings. Nothing else is blue except the piano keys.
- Text hierarchy on `bg`, `bg-band` and `surface`: `ink` for headings and names, `ink-soft` for lead paragraphs, taglines and prose, `ink-muted` for secondary copy and nav links, `ink-subtle` for eyebrows, dates and captions, `ink-faint` only for uppercase labels on `bg`.
- Text on an `accent` fill is `on-accent`.
- Hairlines are `line`. Chip and rail borders are `line-strong`. Image borders are `line-image`. Any border that is the visible edge of a control (outline button, sound toggle, menu button) is `line-control`, which holds 3:1 on `bg`.
- Piano keys have their own tokens (`key-*`). A key is `key-lit` with an `accent` bottom edge when its project was updated in the last 30 days, otherwise `key-idle` with a `key-idle-edge` edge. Black keys are `bg` with a `line-image` outline.

### Type

- Three Google-hosted families: `display` (Fraunces, with the optical-size axis), `sans` (IBM Plex Sans), `mono` (IBM Plex Mono). Load them with one stylesheet link: `https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,400;0,9..144,600;1,9..144,400&family=IBM+Plex+Sans:wght@400;500;600&family=IBM+Plex+Mono:wght@400;500&display=swap`.
- Fraunces carries names and headings only: `hero`, `hero-project`, `page-title`, `section`, `subsection`, `row-title`, `card-title`, `tagline` (italic) and the `wordmark` (italic). Weight 400 for large sizes, 600 for row and card titles.
- Plex Sans carries all reading text: `lead` (21px), `intro` (19px), `prose` (18px, line-height 1.7), `body` (16px), `small` (15px).
- Plex Mono carries metadata only: `eyebrow` (uppercase, 0.08em tracking), `meta` dates, `label`, `chip`. Never set a sentence of prose in mono.
- Below 640px use the `-phone` display sizes (`hero-phone` 68px, `hero-project-phone` 84px, `page-title-phone` 56px, `section-phone` 40px).
- Keep prose under about 65 characters a line: `lead` max 620px, `prose` max 680px.

### Spacing and layout

- Desktop content is `content-width` (1120px) between `gutter` (80px) side paddings: a 1280px frame. Below 640px the gutter is `gutter-phone` (20px) and everything stacks to one column.
- Sections breathe: `space-11` (96px) to `space-12` (112px) between major sections, `space-8` (48px) from a section heading to its content, `space-5` (24px) between cards.
- Two-column heroes use `minmax(0, 1fr)` plus a fixed column (440px on Home for the updates card, 460px on a project page for the screenshot), with `space-10` (80px) or 72px between.
- Lay out siblings with flex or grid and `gap`, never margins between siblings.

### Shape, borders and depth

- Radii: `radius-sm` chips, `radius-md` thumbnails and key ends, `radius-lg` gallery screenshots, `radius-xl` cards and the hero screenshot, `radius-pill` every button and badge.
- Borders, not shadows, separate things. The only shadow on a page is `shadow-lift` on the single hero screenshot of a project page; `shadow-key-pressed` exists for the keyboard only.
- Cards (`surface` with a `line` border) are used for exactly two things: the latest-updates list on Home and hobby cards. Project rows, highlights and meta are flat, divided by `line` rules.

### Motion and interaction

- Keys press in 90ms: white keys dip (`scaleY(0.985)`) and darken, black keys shorten by 4px. `prefers-reduced-motion` removes the transform and transition, keeping only the color change.
- Hovering a key plays its note, and so does tabbing onto a white key (see Octave keyboard). Sound is on by default, remembered per browser, and switchable with the Sound toggle. Browsers stay silent until the visitor's first click or tap anywhere on the page; that is expected.
- Focus: a solid 2px `accent` outline, offset 3px, on every link and button.
- Touch targets are at least `control-height-sm` (44px).

## Iconography

- Inline SVG, 24×24 viewBox, `stroke="currentColor"`, stroke width 1.6 to 2, round caps and joins, no fills except the black keys in the piano glyph. Set icon color through `currentColor` so it follows the text or `accent`.
- The set in use: arrow-right (`M5 12h14M13 6l6 6-6 6`), external link, piano (a rounded rectangle with three key lines and two filled black keys), code brackets with a slash, gamepad, speaker with waves (sound on), speaker with an × (sound off), two-line menu. Draw new icons in the same style; no icon font, no emoji.
- There is no logo file. The brand mark is the `wordmark` style set in type: `bitofbytes` in Fraunces italic.

## Pages

These three page types are the whole site. Build them from the components below.

- **Home.** Header; a two-column hero (eyebrow "Senior Software Engineer · Raleigh, NC", `hero` name with the last name italic in `accent`, `lead` intro, a primary "See the projects" and secondary "Resume (PDF)" button) beside the Update list of the four most recent project updates; the Octave keyboard section ("Eight side projects, one octave" / "What I'm building", with the Sound toggle); the hobbies band (`bg-band`, three Hobby cards: Programming, Piano, Videogames, each linking to the project it inspired and ending in a "now" line); the home footer ("Say hello.", email, GitHub, LinkedIn, location).
- **Projects.** Header; eyebrow "Projects · N", `page-title` "Side *projects*", `intro`; the Sort toggle (Recently updated, Newest first); one Project row per project, sorted by last update, newest first; footer.
- **Project detail.** Header; Breadcrumb; a two-column hero (accent eyebrow with an icon when the project ties to a hobby, `hero-project` name, italic `tagline`, "Open live site" primary and "View on GitHub" secondary buttons, a Started / Last update / Stack meta row) beside the hero screenshot; "What it does" `prose` paragraphs and roman-numeral Highlights beside an aside holding Stack chips and the Timeline; the Screenshots gallery; footer. There is no previous/next project navigation and no separate "latest update" box: the Timeline carries the latest update.

## Components

- **Site header** (`nc-header`): wordmark left, nav right (Home, Projects, Resume, Contact) with `aria-current="page"` on the current item. Below 640px the nav collapses into a 44px round menu button that opens a `<details>` drop-down.
- **Button** (`nc-btn--primary`, `nc-btn--secondary`): pills, 52px tall. One primary per view.
- **Sound toggle** (`nc-sound`): an `aria-pressed` button that switches keyboard audio.
- **Sort toggle** (`nc-sort`): a segmented pair of options, as `aria-pressed` buttons or as links to `?sort=` with `aria-current="true"` on the current one (what the site uses).
- **Update list** (`nc-updates` in an `nc-card`): project name, short date and note per row.
- **Octave keyboard** (`nc-octave`): the signature. Eight white keys, one per project, left to right in the order the projects were started (first commit date, oldest first). A key is lit when that project's last update is within 30 days. White keys play C4 to C5; black keys (C#, D#, F#, G#, A#) are decorative, `aria-hidden`, and play their sharps. Below 640px it turns sideways.
- **Hobby card** (`nc-hobby` in an `nc-card`).
- **Project row** (`nc-row`): thumbnail, name with an optional "Updated recently" Badge, tagline, "Latest ·" note, Tech chips, and an Updated / Started meta column with a Details link.
- **Tech chip** (`nc-chip`) and **Badge** (`nc-badge`).
- **Timeline** (`nc-timeline`): newest first; the newest item has a filled `accent` dot and a "latest" tag; the first commit closes the list.
- **Screenshot** (`nc-shot`, `nc-hero-shot`).
- **Footer** (`nc-footer`, `nc-footer--home`).

## Building bitofbytes with Nocturne

- The site is Go `html/template` plus Tailwind. `tailwind/styles.css` holds the tokens in `:root` and `components/bundle.css` inside `@layer components`; `make tail-prod` builds `static/styles.css` (never edit that file). Page layouts (`nc-home-hero`, `nc-section`, `nc-page-head`, `nc-detail-hero`, `nc-detail-body`) are in the same stylesheet.
- `components/bundle.js` ships as `static/nocturne.js`, loaded with `defer` on Home only. It attaches itself to every `[data-nc-octave]` and `[data-nc-sound]` on the page.
- Project content stays in `models/project.go`; display helpers (dates, the latest note, sorting) are in `models/project_display.go`. "Recently updated" means `LastUpdate` within 30 days of the request (`models.RecentWindow`); keyboard order comes from `FirstCommitDate`; the projects page sorts by `LastUpdate` unless `?sort=newest`.
- The hobby cards' "now" lines come from `models.CurrentActivities()`. An empty field hides its line; the live site never shows a bracketed placeholder.
