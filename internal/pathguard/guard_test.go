package pathguard_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/testutil"
)

func assertReason(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("reason = %q, want it to contain %q", got, want)
	}
}

func TestAcceptsFolderInsideAllowedRoot(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/SomeApp/data.bin", 1)

	if reason := pathguard.New(home).Check(filepath.Join(home, "Library/Caches/SomeApp")); reason != "" {
		t.Fatalf("expected safe, got %q", reason)
	}
}

func TestRejectsRootItself(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/x", 1)

	assertReason(t, pathguard.New(home).Check(filepath.Join(home, "Library/Caches")), "outside")
}

func TestRejectsPathsOutsideAllowedRoots(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Documents/thesis.pdf", 1)
	guard := pathguard.New(home)

	assertReason(t, guard.Check(filepath.Join(home, "Documents")), "outside")
	assertReason(t, guard.Check(home), "outside")
	assertReason(t, guard.Check("/"), "outside")
}

func TestRejectsDotDot(t *testing.T) {
	home := testutil.FakeHome(t)

	assertReason(t, pathguard.New(home).Check(home+"/Library/Caches/../../Documents"), "normalized")
}

func TestRejectsRelativePaths(t *testing.T) {
	assertReason(t, pathguard.New("/home").Check("Library/Caches/x"), "absolute")
}

func TestRejectsSymlink(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Documents/important.txt", 1)
	testutil.WriteFile(t, home, "Library/Caches/keep", 1)
	link := filepath.Join(home, "Library/Caches/evil")
	if err := os.Symlink(filepath.Join(home, "Documents"), link); err != nil {
		t.Fatal(err)
	}

	assertReason(t, pathguard.New(home).Check(link), "symbolic link")
}

func TestRejectsFolderBehindSymlinkedParent(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Documents/project/file.txt", 1)
	testutil.WriteFile(t, home, "Library/Caches/keep", 1)
	if err := os.Symlink(filepath.Join(home, "Documents"), filepath.Join(home, "Library/Caches/link")); err != nil {
		t.Fatal(err)
	}

	assertReason(t, pathguard.New(home).Check(filepath.Join(home, "Library/Caches/link/project")), "resolves outside")
}

func TestAssertSafeReturnsTypedError(t *testing.T) {
	home := testutil.FakeHome(t)

	err := pathguard.New(home).AssertSafe(filepath.Join(home, "Documents"))

	var unsafe *pathguard.UnsafePathError
	if !errors.As(err, &unsafe) || !errors.Is(err, pathguard.ErrUnsafePath) {
		t.Fatalf("expected UnsafePathError, got %v", err)
	}
}

func TestExpandGlob(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/foo-updater/a", 1)
	testutil.WriteFile(t, home, "Library/Caches/bar-updater/a", 1)
	testutil.WriteFile(t, home, "Library/Caches/keep/a", 1)

	got := pathguard.ExpandAll([]string{"~/Library/Caches/*-updater", "~/Library/Caches/missing"}, home)
	slices.Sort(got)

	want := []string{filepath.Join(home, "Library/Caches/bar-updater"), filepath.Join(home, "Library/Caches/foo-updater")}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
