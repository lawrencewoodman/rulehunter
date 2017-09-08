package version

import (
	"testing"
)

func TestVersion_exported(t *testing.T) {
	got := Version()
	if len(got) <= 2 {
		t.Errorf("Version - got: %s", got)
	}
}
func TestVersion_unexported(t *testing.T) {
	cases := []struct {
		major  int
		minor  int
		patch  int
		suffix string
		want   string
	}{
		{major: 0, minor: 5, patch: 0, suffix: "", want: "0.05"},
		{major: 0, minor: 25, patch: 0, suffix: "", want: "0.25"},
		{major: 0, minor: 5, patch: 0, suffix: "-DEV", want: "0.05-DEV"},
		{major: 0, minor: 25, patch: 0, suffix: "-DEV", want: "0.25-DEV"},
		{major: 0, minor: 5, patch: 78, suffix: "", want: "0.05.78"},
		{major: 0, minor: 25, patch: 78, suffix: "", want: "0.25.78"},
		{major: 0, minor: 5, patch: 78, suffix: "-DEV", want: "0.05.78-DEV"},
		{major: 0, minor: 25, patch: 78, suffix: "-DEV", want: "0.25.78-DEV"},
	}
	for i, c := range cases {
		got := version(c.major, c.minor, c.patch, c.suffix)
		if got != c.want {
			t.Errorf("(%02d) version - got: %s, want: %s", i, got, c.want)
		}
	}
}
