# hp-source

A personal knowledge base and static site generator, built as one Go binary. It
tracks what I read (and, over time, other things worth keeping a record of), and
publishes that record alongside my writing on a statically generated site.

## What it is

`hp` is a CLI backed by a single SQLite file. The same binary captures data and
renders the public site from it: it keeps a reading log, holds my writing (longer
articles and short notes), and builds a static site from all of it. Data goes in
through interactive terminal prompts; the site comes out as plain files.

## Ethos

This is a tool for one person, so it is allowed to be small and opinionated.

- **One binary, few dependencies.** No Node, no separate server, no JS framework.
- **Capture should be cheap.** Writing something down should be as quick as a note
  in a margin. Friction is the enemy of keeping a record at all.
- **Simplicity over cleverness.** Plain Go and plain SQL, few packages, comments
  that explain the non-obvious rather than restate the code.
- **Static output.** The published site is just files: fast, durable, and it
  outlives any running process.

## Design influence

The look is drawn from **field journals and travel notes**: the way a notebook is
organised when you are writing things down as you go. The page is divided the way
you would divide a ruled page, with a hairline to separate one entry from the
next, a box to fence off an aside, a single red stroke in the margin where
something is worth marking.

## Conventions

- No em dashes anywhere, code or prose.
- Soft line length around 120; avoid wrapping where you can.
- Comments are tight and name-led, noting only the non-obvious.
- Tests never touch a real database: mock migrations and data, use a temp db.
- Compiled templates and CSS are committed; regenerate them before committing so
  the output does not go stale.
