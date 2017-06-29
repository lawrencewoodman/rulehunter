package main

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	cases := []struct {
		in   []string
		want *cmdFlags
	}{
		{in: []string{"-install"}, want: &cmdFlags{install: true, serve: false}},
		{in: []string{"-install", "-serve"},
			want: &cmdFlags{install: true, serve: true}},
		{in: []string{"-install", "-configdir=/tmp"},
			want: &cmdFlags{install: true, configDir: "/tmp"}},
	}
	for _, c := range cases {
		got := parseFlags(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseFlags - got: %v, want: %v", got, c.want)
		}
	}
}
