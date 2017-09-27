package internal

import "testing"

func TestMakeBuildFilename(t *testing.T) {
	cases := []struct {
		category string
		title    string
	}{
		{category: "", title: "this is the title"},
		{category: "", title: "this is the title "},
		{category: "", title: " this is the title"},
		{category: "", title: "This is the title"},
		{category: "testing", title: "this is the title"},
		{category: "testing", title: "This is the title"},
		{category: "Testing", title: "this is the title"},
		{category: "Testing", title: "This is the title"},
	}
	filenames := map[string]bool{}
	for _, c := range cases {
		filename := MakeBuildFilename(c.category, c.title)
		for i := 0; i < 1000; i++ {
			filename2 := MakeBuildFilename(c.category, c.title)
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
