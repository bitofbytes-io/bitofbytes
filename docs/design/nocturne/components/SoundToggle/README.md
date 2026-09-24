# SoundToggle

Button that turns the Octave keyboard's hover notes on and off.

- Markup: `<button type="button" class="nc-sound" data-nc-sound aria-pressed="true">` holding both speaker icons (`nc-sound__on`, `nc-sound__off`) and a `<span data-nc-sound-label>Sound on</span>`. CSS shows the right icon from `aria-pressed`; `bundle.js` flips the attribute and label and remembers the choice in `localStorage` (`nocturne-sound`).
- Place it beside the keyboard's description, right-aligned on desktop, under the description on phone.
- 44px tall, `line-control` border, `ink-soft` label at 14px/500.
- Sound defaults to on. Browsers keep audio silent until the first click or tap; clicking this toggle also unlocks audio.
