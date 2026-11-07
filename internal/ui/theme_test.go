package ui_test

import (
	"testing"

	"github.com/jefferson-calmon/clean-my-cache/internal/ui"
)

func TestFormatBytes(t *testing.T) {
	cases := map[int64]string{
		512:                        "512 B",
		2048:                       "2 KB",
		5 * 1024 * 1024:            "5.0 MB",
		250 * 1024 * 1024:          "250 MB",
		3 * 1024 * 1024 * 1024 / 2: "1.5 GB",
	}
	for bytes, want := range cases {
		if got := ui.FormatBytes(bytes); got != want {
			t.Errorf("FormatBytes(%d) = %q, want %q", bytes, got, want)
		}
	}
}
