package internal

import (
	"strings"
	"testing"
)

func TestMakeBuildFilename_train(t *testing.T) {
	cases := []struct {
		prefix   string
		category string
		title    string
	}{
		{prefix: "train", category: "", title: "this is the title"},
		{prefix: "train", category: "", title: "this is the title "},
		{prefix: "train", category: "", title: " this is the title"},
		{prefix: "train", category: "", title: "This is the title"},
		{prefix: "train", category: "testing", title: "this is the title"},
		{prefix: "train", category: "testing", title: "This is the title"},
		{prefix: "train", category: "Testing", title: "this is the title"},
		{prefix: "train", category: "Testing", title: "This is the title"},
		{prefix: "test", category: "", title: "this is the title"},
		{prefix: "test", category: "", title: "this is the title "},
		{prefix: "test", category: "", title: " this is the title"},
		{prefix: "test", category: "", title: "This is the title"},
		{prefix: "test", category: "testing", title: "this is the title"},
		{prefix: "test", category: "testing", title: "This is the title"},
		{prefix: "test", category: "Testing", title: "this is the title"},
		{prefix: "test", category: "Testing", title: "This is the title"},
	}
	filenames := map[string]bool{}
	for _, c := range cases {
		filename := MakeBuildFilename(c.prefix, c.category, c.title)
		if !strings.HasPrefix(filename, c.prefix+"_") {
			t.Errorf("MakeBuildFilename - filename: %s, doesn't have prefix: %s_",
				filename, c.prefix)
		}
		for i := 0; i < 1000; i++ {
			filename2 := MakeBuildFilename(c.prefix, c.category, c.title)
			if filename != filename2 {
				t.Errorf(
					"MakeBuildFilename isn't idempotent for category: %s, title: %s",
					c.category,
					c.title,
				)
			}
		}
		filenames[filename] = true
	}
	if len(filenames) != len(cases) {
		t.Error("len(filenames) != len(cases)")
	}
}
