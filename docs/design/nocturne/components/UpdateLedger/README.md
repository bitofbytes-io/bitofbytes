# UpdateLedger

The Home page's "Latest updates" section: one Featured update with a live screenshot beside a flat Update ledger of the next few. It answers "what has Daniel been working on?"

## What the consumer provides

- **Featured update**: the most recently updated project that has a screenshot and is not this site itself (the home page never features a screenshot of itself). Its thumbnail path and alt text, name, short date ("Sep 6"), tagline and latest note.
- **Update ledger**: the next three projects by `LastUpdate`, newest first, leaving out the featured one. BitOfBytes still appears here.
- The recent count for the eyebrow: projects whose `LastUpdate` is within the rolling 30-day `RecentWindow`, not the calendar month. "1 update" or "N updates"; omit the count when zero.
- Notes are the project's `Notes` with the "Most recent work…" lead-in removed (`LatestNote`).

## Markup

```html
<section class="nc-wrap nc-section" aria-labelledby="latest-title">
  <div class="nc-section__titles">
    <p class="nc-eyebrow">Changelog · 6 updates in the last 30 days</p>
    <h2 class="nc-heading nc-heading--section" id="latest-title">Latest <span class="nc-em">updates.</span></h2>
  </div>
  <div class="nc-latest">
    <a class="nc-featured" href="/projects/dined">
      <img class="nc-featured__shot" src="…" alt="…" width="520" height="390" loading="lazy">
      <span class="nc-featured__meta">
        <span class="nc-featured__head"><span class="nc-featured__name">Dined</span><span class="nc-featured__date">Sep 6</span></span>
        <span class="nc-featured__tagline">Proof that nobody actually agreed on dinner.</span>
        <span class="nc-featured__note">Improved visit reliability under load and tightened redirect safety.</span>
      </span>
    </a>
    <div class="nc-ledger">
      <a class="nc-ledger__row" href="/projects/bitofbytes">
        <span class="nc-ledger__date">Sep 24</span>
        <span class="nc-ledger__body"><span class="nc-ledger__name">BitOfBytes</span><span class="nc-ledger__note">…</span></span>
      </a>
      <!-- two more rows -->
    </div>
  </div>
</section>
```

## Look

- Two equal columns (`nc-latest`, 80px gap, 48px below 1024px), one column below 860px with the feature first. With no featured project the ledger spans the full width.
- Featured: a 4:3 screenshot cropped from the top, `radius-lg`, `line-image` border on a `surface` fill; name in Fraunces 600 28px with the short date in `meta` mono at the right; italic Fraunces tagline 20px `ink-soft`; note 15px `ink-muted`. Hovering turns the name `accent`.
- Ledger: flat, no card. Each row is one link: an 88px mono date column (13px `ink-subtle`), then the name (Plex Sans 600 17px `ink`) over the note (15px `ink-muted`), divided by `line` rules. Hovering turns the name `accent`. There is no "see all" link: the nav and the hero button already lead to the projects page.
- Phone (below 640px): name 24px, tagline 18px, note 14px on the feature; ledger rows stack with the name and date on one line and the note underneath, ruled top and bottom.
- Don't: put the ledger in a card, add a live dot, or show more than one screenshot.
