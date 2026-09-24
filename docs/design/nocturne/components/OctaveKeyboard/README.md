# OctaveKeyboard

The signature component: the Home page's project index drawn as one octave of a piano, where every white key is a project and hovering plays its note.

## What the consumer provides

- Exactly eight projects, ordered by `FirstCommitDate`, oldest first (left on desktop, top on phone). If the portfolio grows past eight, show the eight most recently updated in start order and keep "Browse all projects" below.
- For each: the project name, the start label ("since 2024", "since Oct '25"), the link to its detail page, and whether it is lit (`LastUpdate` within 30 days of the build).

## Markup

```html
<div class="nc-octave" data-nc-octave style="--keys: 8">
  <div class="nc-octave__keys">
    <a class="nc-key nc-key--lit" href="/projects/bitofbytes" data-note="C4">
      <span class="nc-key__name">BitOfBytes</span><span class="nc-key__since">since 2024</span>
    </a>
    <!-- … D4 E4 F4 G4 A4 B4 C5; drop nc-key--lit when not updated in 30 days -->
  </div>
  <div class="nc-black" aria-hidden="true" style="--pos:1" data-note="C#4"></div>
  <div class="nc-black" aria-hidden="true" style="--pos:2" data-note="D#4"></div>
  <div class="nc-black" aria-hidden="true" style="--pos:4" data-note="F#4"></div>
  <div class="nc-black" aria-hidden="true" style="--pos:5" data-note="G#4"></div>
  <div class="nc-black" aria-hidden="true" style="--pos:6" data-note="A#4"></div>
</div>
```

- White keys are real links (`<a>`), so the keyboard is a working, keyboard-focusable project index without JavaScript.
- Set `--keys` on `.nc-octave` to the number of white keys rendered (the site writes `len .Keys`); widths, black-key positions and the phone height follow it.
- Black keys are decorative: `aria-hidden`, no link, no focus. Hovering one plays its sharp as a flourish, but they carry no information, so they have no keyboard equivalent; every project is reachable through its white-key link. `--pos` is the white-key boundary they sit on (1, 2, 4, 5, 6); there is none between E and F or after B, and the site omits any whose boundary is past the last key.
- Load `bundle.js` (deferred). It attaches to every `[data-nc-octave]`: pointer-enter adds `is-pressed` and plays the key's `data-note`, pointer-leave releases it. Tabbing onto a white key (keyboard focus, `:focus-visible` only) plays it too. A first touch that arrives before audio is unlocked plays on pointer-down or pointer-up instead of staying silent.

## Look

- Desktop: 1120 × 280px (`key-height`), white keys share the width with 4px gaps, rounded at the bottom (`radius-md`), with a 6px bottom edge (`accent` when lit, `key-idle-edge` otherwise). Name in Plex Sans 600 16px `key-ink`; start label in `chip` mono, `key-ink-lit` or `key-ink-idle`. Black keys 70 × 160px, `bg` with a `line-image` outline.
- Phone (below 640px): the keyboard turns sideways. Keys stack as 60px rows (`key-row-height`) with the name and start label on one line, the edge moves to the left side, and black keys come in from the right as 130 × 36px bars.
- Pressed: lit keys go `key-lit-pressed`, idle keys `key-idle-pressed`, both dip slightly with `shadow-key-pressed`; black keys go `key-black-pressed` and shorten 4px. Reduced motion keeps only the color change.

## Sound

- White keys play C4 to C5 (261.63 to 523.25 Hz), black keys their sharps. The tone is a soft synthesized pluck (triangle plus two quiet partials through a closing low-pass, 1.6s decay); no audio files.
- Pair it with a SoundToggle. Audio stays silent until the visitor's first click or tap; never autoplay, never show a prompt about it.
- Don't: make black keys links, add labels to black keys, animate keys on load, play anything on mouse focus or scroll, or autoplay.
