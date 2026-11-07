package clean_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jefferson-calmon/clean-my-cache/internal/clean"
	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
	"github.com/jefferson-calmon/clean-my-cache/internal/testutil"
)

var updates = recipe.Recipe{ID: "pending-updates", Paths: []string{"~/Library/Caches/*-updater"}}

func setup(t *testing.T, mode clean.Mode) (string, []scan.Item, *clean.Cleaner) {
	t.Helper()
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/app-updater/update.zip", 2*testutil.MiB)
	guard := pathguard.New(home)
	runner := &testutil.FakeRunner{}

	items := (&scan.Scanner{Home: home, Runner: runner, Guard: guard}).Scan(context.Background(), recipe.Set{updates})
	now := func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	return home, items, &clean.Cleaner{Home: home, Runner: runner, Guard: guard, Mode: mode, Now: now}
}

func TestDeletesOnlyScannedFolders(t *testing.T) {
	home, items, cleaner := setup(t, clean.ModeDelete)
	keep := testutil.WriteFile(t, home, "Library/Caches/keep-me/data", 1)

	outcome := cleaner.Clean(context.Background(), items[0])

	if !outcome.OK || outcome.FreedBytes < 2*testutil.MiB {
		t.Fatalf("outcome = %+v", outcome)
	}
	if testutil.Exists(filepath.Join(home, "Library/Caches/app-updater")) {
		t.Fatal("cache folder still exists")
	}
	if !testutil.Exists(keep) {
		t.Fatal("unrelated folder was deleted")
	}
}

func TestMovesToTrash(t *testing.T) {
	home, items, cleaner := setup(t, clean.ModeTrash)

	cleaner.Clean(context.Background(), items[0])

	entries, err := os.ReadDir(filepath.Join(home, ".Trash"))
	if err != nil || len(entries) != 1 || !strings.HasPrefix(entries[0].Name(), "app-updater ") {
		t.Fatalf("trash = %v, err = %v", entries, err)
	}
}

func TestRefusesPathSwappedForSymlinkAfterScan(t *testing.T) {
	home, items, cleaner := setup(t, clean.ModeDelete)
	target := filepath.Join(home, "Library/Caches/app-updater")
	precious := testutil.WriteFile(t, home, "Documents/precious.txt", 1)
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Dir(precious), target); err != nil {
		t.Fatal(err)
	}

	outcome := cleaner.Clean(context.Background(), items[0])

	if outcome.OK || !testutil.Exists(precious) {
		t.Fatalf("expected refusal and intact documents, got %+v", outcome)
	}
}
