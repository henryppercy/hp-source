package site

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// devAssetDir is where templates and static assets live on disk, read by
// serve --watch so live edits are reflected without recompiling the binary.
const devAssetDir = "internal/site"

// Serve builds the site and serves it over HTTP. With watch enabled it reads
// templates and assets from disk and rebuilds when they change.
func Serve(out, addr string, watch bool) error {
	assets := embeddedAssets()
	if watch {
		assets = os.DirFS(devAssetDir)
	}

	b := newBuilder(assets, out)
	if err := b.Build(); err != nil {
		return err
	}

	if watch {
		go watchAndRebuild(b)
	}

	log.Printf("serving %s on http://localhost%s", out, addr)
	return http.ListenAndServe(addr, fileServer(out))
}

// fileServer serves the built site, falling back to the generated 404.html
// (instead of the plain-text default) for paths that do not exist.
func fileServer(out string) http.Handler {
	fs := http.FileServer(http.Dir(out))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(&notFoundWriter{ResponseWriter: w, out: out}, r)
	})
}

type notFoundWriter struct {
	http.ResponseWriter
	out     string
	rewrote bool
}

func (w *notFoundWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		if data, err := os.ReadFile(filepath.Join(w.out, "404.html")); err == nil {
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

func watchAndRebuild(b *builder) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watch: %v", err)
		return
	}
	defer w.Close()

	for _, dir := range []string{
		filepath.Join(devAssetDir, "templates"),
		filepath.Join(devAssetDir, "static"),
	} {
		if err := w.Add(dir); err != nil {
			log.Printf("watch %s: %v", dir, err)
		}
	}

	var timer *time.Timer
	for {
		select {
		case _, ok := <-w.Events:
			if !ok {
				return
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(100*time.Millisecond, func() {
				if err := b.Build(); err != nil {
					log.Printf("rebuild: %v", err)
					return
				}
				log.Println("rebuilt")
			})
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("watch: %v", err)
		}
	}
}
