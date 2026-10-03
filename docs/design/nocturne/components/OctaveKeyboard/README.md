# OctaveKeyboard

The signature component: the Home page's project index drawn as one octave of a piano, where every white key is a project and hovering plays its note. It sits directly under the hero, with no heading of its own: a caption and legend under the keys say what it is.

## What the consumer provides

- Exactly eight projects, ordered by `FirstCommitDate`, oldest first (left on desktop, top on phone). If the portfolio grows past eight, show the eight most recently updated in start order; the Latest updates section links to every project.
- For each: the project name, the start label ("since 2024", "since Oct '25"), the link to its detail page, and whether it is lit (`LastUpdate` within 30 days of the build).

## Markup

```html
<div class="nc-octave nc-octave--keys-8" data-nc-octave>
  <div class="nc-octave__keys">
    <a class="nc-key nc-key--lit" href="/projects/bitofbytes" data-note="C4">
      <span class="nc-key__name">BitOfBytes</span><span class="nc-key__since">since 2024</span>
    </a>
    <!-- … D4 E4 F4 G4 A4 B4 C5; drop nc-key--lit when not updated in 30 days -->
  </div>
  <div class="nc-black nc-black--pos-1" aria-hidden="true"></div>
  <div class="nc-black nc-black--pos-2" aria-hidden="true"></div>
  <div class="nc-black nc-black--pos-4" aria-hidden="true"></div>
  <div class="nc-black nc-black--pos-5" aria-hidden="true"></div>
  <div class="nc-black nc-black--pos-6" aria-hidden="true"></div>
</div>
```

- White keys are real links (`<a>`), so the keyboard is a working, keyboard-focusable project index without JavaScript.
- Add `nc-octave--keys-N` to `.nc-octave` for the number of white keys rendered, 1 to 8 (the site writes `len .Keys`); it sets `--keys`, and widths, black-key positions and the phone height follow it. Each black key takes `nc-black--pos-N` for its boundary. The site's CSP blocks inline styles, so never set these with a `style` attribute.
- Black keys are purely decorative: `aria-hidden`, no link, no focus, no sound, and `pointer-events: none`, so hovering one plays the white key underneath. Every sound on the keyboard is therefore reachable by keyboard too. `--pos` is the white-key boundary they sit on (1, 2, 4, 5, 6); there is none between E and F or after B, and the site omits any whose boundary is past the last key.
- Load `static/nocturne.js` (deferred). It attaches to every `[data-nc-octave]`: pointer-enter adds `is-pressed` and plays the key's `data-note`, pointer-leave releases it. Tabbing onto a white key (keyboard focus, `:focus-visible` only) plays it too. A first touch that arrives before audio is unlocked plays on pointer-down or pointer-up instead of staying silent.

## Caption and legend

Under the keys, one row (`nc-octave__caption`): the legend on the left, the SoundToggle on the right. It stacks below 860px.

```html
<section class="nc-wrap nc-octave-section" aria-label="Eight side projects, one octave">
  <div class="nc-octave nc-octave--keys-8" data-nc-octave>…</div>
  <div class="nc-octave__caption">
    <p class="nc-legend">
      <span>Eight side projects, one octave. <span class="nc-desktop-only">Left to right</span><span class="nc-phone-only">Top to bottom</span> in the order I started them.</span>
      <span class="nc-legend__item"><span class="nc-legend__swatch nc-legend__swatch--lit" aria-hidden="true"></span>updated in the last 30 days</span>
      <span class="nc-legend__item"><span class="nc-legend__swatch" aria-hidden="true"></span>resting</span>
    </p>
    <button type="button" class="nc-sound" data-nc-sound aria-pressed="true">
      <svg class="nc-sound__on" …></svg><svg class="nc-sound__off" …></svg>
      <span><span class="nc-desktop-only">Hover to play · </span><span class="nc-phone-only">Tap to play · </span><span data-nc-sound-label>Sound on</span></span>
    </button>
  </div>
</section>
```

- The legend is 15px `ink-subtle` (14px on phone). Each swatch is a 14px miniature key: `key-lit` with a 3px `accent` bottom edge, or `key-idle` with a `key-idle-edge` edge. On phone the edge moves to the left, like the sideways keys.
- `static/nocturne.js` only rewrites the `[data-nc-sound-label]` text, so the "Hover to play ·" prefix sits outside it. The extra wrapping span keeps the prefix and the label together inside the button's flex gap.

## Look

- Desktop: 1120 × 220px (`key-height`), white keys share the width with 4px gaps, rounded at the bottom (`radius-md`), with a 6px bottom edge (`accent` when lit, `key-idle-edge` otherwise). Name in Plex Sans 600 16px `key-ink`; start label in `chip` mono, `key-ink-lit` or `key-ink-idle`. Black keys 70 × 126px (`key-black-height`, half a key slot wide and 57% of the white key), `bg` with a `line-image` outline.
- Phone (below 640px): the keyboard turns sideways. Keys stack as 60px rows (`key-row-height`) with the name and start label on one line, the edge moves to the left side, and black keys come in from the right as 130 × 36px bars.
- Lit keys (`key-lit`) are visibly bluer than idle ones (`key-idle`) at a glance; the legend names both.
- Pressed: lit keys go `key-lit-pressed`, idle keys `key-idle-pressed`, both dip slightly with `shadow-key-pressed`; Reduced motion keeps only the color change.

## Sound

- White keys play C4 to C5 (261.63 to 523.25 Hz); black keys are silent. The tone is a soft synthesized pluck (triangle plus two quiet partials through a closing low-pass, 1.6s decay); no audio files.
- Pair it with a SoundToggle. Audio stays silent until the visitor's first click or tap; never autoplay, never show a prompt about it.
- Don't: make black keys links, add labels to black keys, animate keys on load, play anything on mouse focus or scroll, or autoplay.
