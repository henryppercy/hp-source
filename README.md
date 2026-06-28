wip: personal data management tool

## Build

Templates are written in [templ](https://templ.guide) and CSS is built with the
Tailwind standalone CLI (no Node). templ and air are pinned as Go tool directives,
so no install step is needed beyond the Tailwind binary. One-time:

```
make tailwind   # download the pinned standalone binary to ./bin
```

Then:

```
make css        # generate internal/site/static/styles/app.css (committed)
make generate   # compile *.templ to *_templ.go (committed)
make build      # css + generate + go build -o bin/hp
```

## Dev

Run two watchers, each in its own terminal:

```
make css-watch  # terminal 1: rebuild app.css on save
make dev        # terminal 2: air (regenerate templ, rebuild, restart serve)
```

air rebuilds and restarts the server on `.go`/`.templ` edits; images hot-reload
through `hp site serve --watch` without a restart. Templates compile into the
binary, so a template edit needs the rebuild that air handles.

The Tailwind watcher has to run in its own foreground terminal: backgrounding it
(stdin not a terminal) stops it from rebuilding, so it can't be folded into `make
dev`.
