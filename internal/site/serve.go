package site

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/henryppercy/hp-source/internal/repo"
)

// devAssetDir is where static assets live on disk, read by serve --watch so CSS
// and image edits are reflected without recompiling. Templates are compiled into
// the binary by templ, so template edits need a rebuild (air handles that in dev).
const devAssetDir = "internal/site"

// ServeConfig tunes the dev server built by Serve.
type ServeConfig struct {
	Addr   string // listen address, e.g. ":8080"
	Watch  bool   // rebuild when static assets or the database change
	Drafts bool   // include draft posts in the build
	DBPath string // database file to watch when Watch is set
}

// Serve builds the site and serves it over HTTP. With watch enabled it reads
// static assets from disk and rebuilds when they or the database change.
func Serve(r *repo.Repo, out string, cfg ServeConfig) error {
	assets := embeddedAssets()
	if cfg.Watch {
		assets = os.DirFS(devAssetDir)
	}

	// Local serving always includes images so the site works offline.
	b := newBuilder(r, assets, out, true, cfg.Drafts)
	if err := b.Build(); err != nil {
		return err
	}

	if cfg.Watch {
		go watchAndRebuild(b, cfg.DBPath)
	}

	log.Info("serving", "out", out, "url", "http://localhost"+cfg.Addr)
	return http.ListenAndServe(cfg.Addr, fileServer(out))
}

// fileServer serves the built site, falling back to the generated 404.html
// (instead of the plain-text default) for paths that do not exist.
func fileServer(out string) http.Handler {
	fs := http.FileServer(http.Dir(out))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(&notFoundWriter{ResponseWriter: w, out: out}, r)
	})
}

// notFoundWriter intercepts a 404 from the file server and substitutes the
// generated 404.html, swallowing the default plain-text body.
type notFoundWriter struct {
	http.ResponseWriter
	out     string
	rewrote bool
}

func (w *notFoundWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		if data, err := os.ReadFile(filepath.Join(w.out, "404.html")); err != nil {
			log.Debug("no 404 page to serve", "error", err)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.ResponseWriter.WriteHeader(http.StatusNotFound)
			w.ResponseWriter.Write(data)
			w.rewrote = true
			return
		}
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *notFoundWriter) Write(b []byte) (int, error) {
	if w.rewrote {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func watchAndRebuild(b *builder, dbPath string) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("watch", "error", err)
		return
	}
	defer w.Close()

	// fsnotify is not recursive, so add every directory under the static root
	// (static/ and its styles/ and images/ subdirs).
	for _, root := range []string{
		filepath.Join(devAssetDir, "static"),
	} {
		if err := fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			dir := filepath.Join(root, path)
			if err := w.Add(dir); err != nil {
				log.Error("watch dir", "dir", dir, "error", err)
			}
			return nil
		}); err != nil {
			// A failed walk leaves the watcher silently incomplete, so
			// say so instead of rebuilding nothing forever.
			log.Error("watch walk", "root", root, "error", err)
		}
	}

	// The content lives in SQLite. In WAL mode commits write to the -wal
	// file, so watch the database's directory rather than the file itself.
	if dbPath != "" {
		if err := w.Add(filepath.Dir(dbPath)); err != nil {
			log.Error("watch db", "dir", filepath.Dir(dbPath), "error", err)
		}
	}

	var timer *time.Timer
	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if !relevantEvent(ev.Name, dbPath) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(100*time.Millisecond, func() {
				if err := b.Build(); err != nil {
					log.Error("rebuild", "error", err)
					return
				}
				// The built summary comes from Build itself.
				log.Debug("rebuilt")
			})
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Error("watch", "error", err)
		}
	}
}

// relevantEvent reports whether a filesystem event should trigger a rebuild.
// Changes outside the database's directory always count. Inside it, only the
// database and its WAL or journal files do.
func relevantEvent(name, dbPath string) bool {
	if dbPath == "" || filepath.Dir(name) != filepath.Dir(dbPath) {
		return true
	}
	// The -shm file moves on ordinary reads, so reacting to it would rebuild
	// for nothing.
	base := filepath.Base(dbPath)
	switch filepath.Base(name) {
	case base, base + "-wal", base + "-journal":
		return true
	}
	return false
}
