> [!NOTE]
> Work in progress 

# HP Source

Personal data management tool and SSG.

## Install

Only install needed is the Tailwind binary:

```
make tailwind   # download the pinned binary to ./bin
```

## Build

```
make build      # builds the app and installs to ~/.local/bin/hp
```

Override with `make build INSTALL_DIR=/some/dir`.

This uses GNU coreutils `install -D`, so only works on Linux. 

## Dev

Run two watchers:

```
make css-watch  # rebuild app.css on save
make dev        # air (regenerate templ, rebuild, restart serve)
```

