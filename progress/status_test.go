package progress

import "testing"

func TestStatusKindString(t *testing.T) {
	cases := []struct {
		status StatusKind
		want   string
	}{
		{status: Waiting, want: "waiting"},
		{status: Processing, want: "processing"},
		{status: Success, want: "success"},
		{status: Error, want: "error"},
	}
	for _, c := range cases {
		got := c.status.String()
		if got != c.want {
			t.Errorf("String() got: %s, want: %s", got, c.want)
		}
	}
}
