package main

import (
	"errors"
	"testing"
)

func TestErrConfigLoadError(t *testing.T) {
	err := errConfigLoad{
		filename: "/tmp/config.yaml",
		err:      errors.New("baby did a bad thing"),
	}
	want :=
		"couldn't load configuration file: /tmp/config.yaml: baby did a bad thing"
	got := err.Error()
	if got != want {
		t.Errorf("Error() got: %s, want: %s", got, want)
	}
}
