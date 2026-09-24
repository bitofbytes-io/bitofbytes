# Screenshot

How project screenshots are framed: one lifted hero shot, then a flat two-up gallery.

- **Hero** (`nc-hero-shot`): the right column of a project page's hero, 460 × 614px, `object-fit: cover`, `radius-xl`, `line-image` border and `shadow-lift`. The only shadow on the page. Full width at 3:4 on phone.
- **Gallery** (`nc-shots` of `figure.nc-shot`): two columns, 24px gap, images cropped from the top at 548:520, `radius-lg`, `line-image` border. The caption row puts the screenshot title (600, `ink`) left and its note (`ink-subtle`) right.
- Every image needs real `alt` text from the project's screenshot data. Screenshots link to the full-size file.
