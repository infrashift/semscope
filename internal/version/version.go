// Package version holds the build metadata of semscope and the SARIF versions
// that it can write.
package version

import (
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
)

// The build task sets these values with -ldflags "-X <pkg>.Version=...".
// The linker sets only package-level string variables, so they are not constants.
var (
	Version   = "" //nolint:gochecknoglobals // set by the linker through -ldflags -X
	Commit    = "" //nolint:gochecknoglobals // set by the linker through -ldflags -X
	BuildDate = "" //nolint:gochecknoglobals // set by the linker through -ldflags -X
)

// The SARIF versions that semscope writes. SARIF 2.2 is an OASIS draft.
const (
	SARIF210            = "2.1.0"
	SARIF22             = "2.2"
	DefaultSARIFVersion = SARIF210
	unknown             = "unknown"
	develVersion        = "dev"
)

// Info is the resolved build metadata of one semscope binary.
type Info struct {
	Version       string
	Commit        string
	BuildDate     string
	GoVersion     string
	SARIFVersions []string
}

// Linker holds the values that -ldflags sets. An empty field is not set.
type Linker struct {
	Version   string
	Commit    string
	BuildDate string
}

// Get returns the build metadata of the running binary.
func Get() Info {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		bi = nil
	}
	info := Resolve(Linker{Version: Version, Commit: Commit, BuildDate: BuildDate}, bi)
	assert.That(info.Commit != "", "Resolve sets the commit")
	assert.That(info.BuildDate != "", "Resolve sets the build date")
	return info
}

// Resolve merges the linker values with the build info bi, which can be nil.
// A linker value has priority. Each empty field gets "unknown" or "dev".
func Resolve(ld Linker, bi *debug.BuildInfo) Info {
	info := Info{
		Version:       ld.Version,
		Commit:        ld.Commit,
		BuildDate:     ld.BuildDate,
		GoVersion:     runtime.Version(),
		SARIFVersions: SARIFVersions(),
	}
	if bi != nil {
		applyBuildInfo(&info, bi)
	}
	if info.Version == "" {
		info.Version = develVersion
	}
	if info.Commit == "" {
		info.Commit = unknown
	}
	if info.BuildDate == "" {
		info.BuildDate = unknown
	}
	assert.That(info.Version != "" && info.Commit != "" && info.BuildDate != "", "each field has a value")
	assert.That(info.GoVersion != "", "the Go version has a value")
	assert.That(len(info.SARIFVersions) == 2, "two SARIF versions")
	return info
}

// applyBuildInfo fills each empty field of info from the module and VCS data of bi.
func applyBuildInfo(info *Info, bi *debug.BuildInfo) {
	assert.That(info != nil, "info is not nil")
	assert.That(bi != nil, "bi is not nil")
	if info.Version == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		info.Version = bi.Main.Version
	}
	if bi.GoVersion != "" {
		info.GoVersion = bi.GoVersion
	}
	revision, modified, vcsTime := "", false, ""
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		case "vcs.time":
			vcsTime = s.Value
		}
	}
	if info.Commit == "" && revision != "" {
		info.Commit = revision
		if modified {
			info.Commit += "-dirty"
		}
	}
	if info.BuildDate == "" {
		info.BuildDate = vcsTime
	}
	assert.That(info.Version != "(devel)", "the devel marker never becomes the version")
}

// SARIFVersions returns the SARIF versions that semscope writes, oldest first.
func SARIFVersions() []string {
	out := []string{SARIF210, SARIF22}
	assert.That(out[0] == DefaultSARIFVersion, "the default SARIF version comes first")
	assert.That(len(out) == 2, "two SARIF versions")
	return out
}

// IsSupportedSARIF reports whether semscope writes the SARIF version v.
func IsSupportedSARIF(v string) bool {
	ok := v == SARIF210 || v == SARIF22
	assert.That(!ok || len(v) >= 3, "a supported version has a major and a minor part")
	return ok
}

// String gives the version text that `semscope --version` prints.
// The text has a fixed field sequence, so the output is the same on each run.
func (i Info) String() string {
	assert.That(i.Version != "", "the Info value comes from Resolve")
	var b strings.Builder
	b.WriteString("semscope " + i.Version + "\n")
	b.WriteString("commit: " + i.Commit + "\n")
	b.WriteString("build date: " + i.BuildDate + "\n")
	b.WriteString("go: " + i.GoVersion + "\n")
	b.WriteString("sarif: " + strings.Join(i.SARIFVersions, ", ") + "\n")
	out := b.String()
	assert.That(strings.Count(out, "\n") == 5, "the version text has five lines")
	return out
}
