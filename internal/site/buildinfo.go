package site

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
)

// RecordBuild resolves the location the build is filed from, prepares the
// footer's build info, and logs a pending build row. It errors before writing
// if the location is unknown. Mark the returned id successful once the build
// completes with MarkBuildSuccess.
func RecordBuild(r *repo.Repo, fromSlug string) (int, error) {
	loc, err := r.GetLocationBySlug(fromSlug)
	if err != nil {
		return 0, fmt.Errorf("build location %q: %w", fromSlug, err)
	}

	templates.LastBuild = templates.BuildInfo{
		Date:     time.Now(),
		Go:       goVersion(),
		On:       builtOn(),
		Location: placeOf(*loc),
	}

	return r.AddBuild(&repo.BuildInput{
		LocationID: loc.ID,
		GoVersion:  templates.LastBuild.Go,
		BuiltOn:    templates.LastBuild.On,
	})
}

// goVersion is the version the project targets, read from the go.mod go
// directive (e.g. "1.25.8"). Outside the repo it falls back to the toolchain
// version, dropping any GOEXPERIMENT suffix like "-X:nodwarf5".
func goVersion() string {
	if v := goModVersion(); v != "" {
		return v
	}
	v := strings.TrimPrefix(runtime.Version(), "go")
	v, _, _ = strings.Cut(v, "-")
	return v
}

// goModVersion returns the go directive from go.mod in the working directory,
// or "" when it cannot be read (e.g. a binary run outside the repo).
func goModVersion() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// builtOn is the OS, kernel release and machine, e.g. "Linux 7.0.12-arch1-1;
// x86_64", falling back to the Go runtime when uname is unavailable.
func builtOn() string {
	if out, err := exec.Command("uname", "-srm").Output(); err == nil {
		if f := strings.Fields(string(out)); len(f) == 3 {
			return fmt.Sprintf("%s %s; %s", f[0], f[1], f[2])
		}
	}
	return fmt.Sprintf("%s; %s", runtime.GOOS, runtime.GOARCH)
}

// liveBuildInfo derives build info for dev serving, stamped now and filed from
// home, used when no build was recorded by the build command.
func liveBuildInfo() templates.BuildInfo {
	return templates.BuildInfo{
		Date:     time.Now(),
		Go:       goVersion(),
		On:       builtOn(),
		Location: templates.HomeLocation,
	}
}
