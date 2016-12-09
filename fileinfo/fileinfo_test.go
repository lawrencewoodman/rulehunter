package fileinfo

import (
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"testing"
	"time"
)

func TestIsEqual(t *testing.T) {
	cases := []struct {
		a    FileInfo
		b    FileInfo
		want bool
	}{
		{a: testhelpers.NewFileInfo(
			"hello.txt",
			testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
		),
			b: testhelpers.NewFileInfo(
				"hello.txt",
				testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
			),
			want: true,
		},
		{a: testhelpers.NewFileInfo(
			"hllo.txt",
			testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
		),
			b: testhelpers.NewFileInfo(
				"hello.txt",
				testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
			),
			want: false,
		},
		{a: testhelpers.NewFileInfo(
			"hello.txt",
			testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
		),
			b: testhelpers.NewFileInfo(
				"helo.txt",
				testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
			),
			want: false,
		},
		{a: testhelpers.NewFileInfo(
			"hello.txt",
			testhelpers.MustParse(time.RFC822, "02 Jan 16 11:21 GMT"),
		),
			b: testhelpers.NewFileInfo(
				"hello.txt",
				testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
			),
			want: false,
		},
		{a: testhelpers.NewFileInfo(
			"hello.txt",
			testhelpers.MustParse(time.RFC822, "02 Jan 16 11:20 GMT"),
		),
			b: testhelpers.NewFileInfo(
				"hello.txt",
				testhelpers.MustParse(time.RFC822, "02 Jan 16 11:21 GMT"),
			),
			want: false,
		},
	}
	for _, c := range cases {
		if got := IsEqual(c.a, c.b); got != c.want {
			t.Errorf("IsEqual(%s, %s) got: %t, want: %t", c.a, c.b, got, c.want)
		}
	}
}
