# Screenshot

How project screenshots are framed: one lifted hero shot, then a flat two-up gallery.

- **Hero** (`nc-hero-shot`): the right column of a project page's hero, 460 × 614px, `object-fit: cover`, `radius-xl`, `line-image` border and `shadow-lift`. The only shadow on the page. Full width at 3:4 on phone.
- **Gallery** (`nc-shots` of `figure.nc-shot`): two columns, 24px gap, images cropped from the top at 548:520, `radius-lg`, `line-image` border. The caption row puts the screenshot title (600, `ink`) left and its note (`ink-subtle`) right.
- Every image needs real `alt` text from the project's screenshot data. Screenshots link to the full-size file, so they work without JavaScript.
- **Viewer** (`nc-viewer`): `static/screenshot-viewer.js`, loaded (deferred) on the projects pages, turns the gallery links (inside `[data-nc-shots]`) into a native modal `<dialog>` that it builds on first load. It shows the full image on a dark scrim with its title, note and "N of M" count, close, previous and next buttons (44px round, `line-control` border), and wraps around. The browser supplies the focus trap, Escape to close and focus return to the clicked link; the script adds the left and right arrow keys and a click on the backdrop to close. Ctrl, Cmd, Shift and middle clicks still open the plain link. On phone the previous and next buttons sit under the image.
