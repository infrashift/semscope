package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

func buildInfo(mainVersion string, settings ...debug.BuildSetting) *debug.BuildInfo {
	return &debug.BuildInfo{
		GoVersion: "go1.99.0",
		Main:      debug.Module{Path: "github.com/infrashift/semscope", Version: mainVersion},
		Settings:  settings,
	}
}

func TestResolve(t *testing.T) {
	rev := debug.BuildSetting{Key: "vcs.revision", Value: "abc123"}
	vcsTime := debug.BuildSetting{Key: "vcs.time", Value: "2026-11-17T00:00:00Z"}
	dirty := debug.BuildSetting{Key: "vcs.modified", Value: "true"}
	tests := []struct {
		name string
		ld   Linker
		bi   *debug.BuildInfo
		want Info
	}{
		{"no data", Linker{}, nil, Info{Version: "dev", Commit: "unknown", BuildDate: "unknown"}},
		{"linker wins", Linker{"v1.2.3", "def456", "2026-01-01T00:00:00Z"}, buildInfo("v0.0.1", rev, vcsTime),
			Info{Version: "v1.2.3", Commit: "def456", BuildDate: "2026-01-01T00:00:00Z", GoVersion: "go1.99.0"}},
		{"build info fallback", Linker{}, buildInfo("v0.4.0", rev, vcsTime),
			Info{Version: "v0.4.0", Commit: "abc123", BuildDate: "2026-11-17T00:00:00Z", GoVersion: "go1.99.0"}},
		{"devel and dirty", Linker{}, buildInfo("(devel)", rev, dirty),
			Info{Version: "dev", Commit: "abc123-dirty", BuildDate: "unknown", GoVersion: "go1.99.0"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.ld, tc.bi)
			if got.Version != tc.want.Version || got.Commit != tc.want.Commit || got.BuildDate != tc.want.BuildDate {
				t.Fatalf("Resolve() = %+v, want %+v", got, tc.want)
			}
			if tc.want.GoVersion != "" && got.GoVersion != tc.want.GoVersion {
				t.Fatalf("GoVersion = %q, want %q", got.GoVersion, tc.want.GoVersion)
			}
			if strings.Join(got.SARIFVersions, ",") != "2.1.0,2.2" {
				t.Fatalf("SARIFVersions = %v", got.SARIFVersions)
			}
		})
	}
}

func TestGetHasEveryField(t *testing.T) {
	got := Get()
	if got.Version == "" || got.Commit == "" || got.BuildDate == "" || got.GoVersion == "" {
		t.Fatalf("Get() = %+v, want every field set", got)
	}
}

func TestIsSupportedSARIF(t *testing.T) {
	for v, want := range map[string]bool{"2.1.0": true, "2.2": true, "2.1": false, "": false, "3.0": false} {
		if got := IsSupportedSARIF(v); got != want {
			t.Errorf("IsSupportedSARIF(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestStringIsDeterministic(t *testing.T) {
	info := Resolve(Linker{"v1.0.0", "abc", "2026-11-17T00:00:00Z"}, buildInfo("", debug.BuildSetting{}))
	want := "semscope v1.0.0\ncommit: abc\nbuild date: 2026-11-17T00:00:00Z\ngo: go1.99.0\nsarif: 2.1.0, 2.2\n"
	if got := info.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
