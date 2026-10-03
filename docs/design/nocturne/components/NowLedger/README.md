# NowLedger

The "Off the clock, *mostly keys.*" band on Home: three flat columns, Programming, Piano, Videogames, each saying what Daniel is doing with it now.

- Sits in an `nc-band` (`bg-band`, `line` top and bottom, 96px padding; 64px on phone) under a `section` heading.
- `nc-now` is three equal columns with no gap; each `nc-now__item` is padded 40px inline (28px below 1024px) and divided from the previous one by a `line` left border. No cards, no icons.
- Each item: an uppercase `label` (`nc-now__label`, mono 12px 500, `ink-faint`) naming the hobby and, when there is a "now" value, what it is ("Programming · building now", "Piano · on the stand", "Videogames · playing"); the value itself in italic Fraunces 30px `ink` (`nc-now__value`, 26px on phone), set as a sentence start ("This redesign"); then one or two sentences of `ink-muted` 16px copy (`nc-now__copy`).
- Piano links to Noted and Videogames to Anthology with `nc-prose-link`.
- The values come from `models.CurrentActivities()`, written with the first letter capitalized as displayed. An empty value hides both the label suffix and the value line. The live site never shows a bracketed placeholder and never invents a piece or a game.
- Below 860px the items stack, ruled by `line` above each and below the last, 20px vertical padding.
- Exactly three items: no fourth hobby without Daniel adding one.
