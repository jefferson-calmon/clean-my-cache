package testutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jefferson-calmon/clean-my-cache/internal/system"
)

func FakeHome(t *testing.T) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return home
}

func WriteFile(t *testing.T, home, relative string, size int) string {
	t.Helper()
	path := filepath.Join(home, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const MiB = 1024 * 1024

type FakeRunner struct {
	Installed map[string]bool
	Outputs   map[string]system.Result
	Calls     []string
}

func (f *FakeRunner) Run(_ context.Context, argv []string) system.Result {
	key := strings.Join(argv, " ")
	f.Calls = append(f.Calls, key)
	return f.Outputs[key]
}

func (f *FakeRunner) HasCommand(bin string) bool {
	return f.Installed[bin]
}

func Exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
