wip: personal data management tool

## Build

CSS is built with the Tailwind standalone CLI (no Node). One-time:

```
make tailwind   # download the pinned standalone binary to ./bin
```

Then:

```
make css        # generate internal/site/static/styles/app.css (committed)
make build      # css + go build -o bin/hp
```

Dev: run `make css-watch` alongside `hp site serve --watch`.
