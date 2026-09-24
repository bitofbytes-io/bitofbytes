# SiteHeader

The top bar of every page: the wordmark on the left, four links on the right.

- Wordmark: `bitofbytes` in the `wordmark` style (Fraunces italic 26px, `ink`), linking home. No logo image.
- Nav: Home, Projects, Resume (the PDF), Contact (the footer anchor on Home). `ink-muted` links; the current section gets `aria-current="page"`, which paints it `accent` at weight 500. A project detail page marks Projects as current.
- 28px vertical padding and a `line` bottom rule, inside `nc-wrap`.
- Below 640px the nav hides and a `<details class="nc-menu">` appears: its `<summary class="nc-menu-btn" aria-label="Menu">` is the round 44px button (`line-control` border) and its `nav.nc-menu__panel` drops down the same four links on `surface`. It works without JavaScript; `static/nav-menu.js` closes it on outside click and Escape.
