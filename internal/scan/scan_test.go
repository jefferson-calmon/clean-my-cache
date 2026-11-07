package scan_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
	"github.com/jefferson-calmon/clean-my-cache/internal/system"
	"github.com/jefferson-calmon/clean-my-cache/internal/testutil"
)

var (
	slack   = recipe.Recipe{ID: "slack", App: "Slack", Risk: recipe.RiskSafe, Paths: []string{"~/Library/Application Support/Slack/Cache"}}
	updates = recipe.Recipe{ID: "pending-updates", Risk: recipe.RiskSafe, Paths: []string{"~/Library/Caches/*-updater"}}
)

func newScanner(home string, runner *testutil.FakeRunner, running map[string]bool) *scan.Scanner {
	return &scan.Scanner{Home: home, Runner: runner, Guard: pathguard.New(home), RunningApps: running}
}

func TestMeasuresCachesAndIgnoresTinyOnes(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/app-updater/update.zip", 2*testutil.MiB)
	testutil.WriteFile(t, home, "Library/Application Support/Slack/Cache/tiny", 10)

	items := newScanner(home, &testutil.FakeRunner{}, nil).Scan(context.Background(), recipe.Set{slack, updates})

	if items[0].Status != scan.StatusEmpty {
		t.Errorf("slack status = %s, want empty", items[0].Status)
	}
	if items[1].Status != scan.StatusReady || items[1].Bytes < 2*testutil.MiB {
		t.Errorf("updates = %s %d bytes, want ready with 2 MiB", items[1].Status, items[1].Bytes)
	}
}

func TestSkipsOpenApps(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Application Support/Slack/Cache/blob", 2*testutil.MiB)

	items := newScanner(home, &testutil.FakeRunner{}, map[string]bool{"Slack": true}).Scan(context.Background(), recipe.Set{slack})

	if items[0].Status != scan.StatusAppRunning {
		t.Fatalf("status = %s, want app-running", items[0].Status)
	}
}

func TestEachFolderIsClaimedOnce(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/Caches/app-updater/update.zip", 2*testutil.MiB)
	specific := recipe.Recipe{ID: "app", Paths: []string{"~/Library/Caches/app-updater"}}

	items := newScanner(home, &testutil.FakeRunner{}, nil).Scan(context.Background(), recipe.Set{specific, updates})

	if len(items[0].Paths) != 1 || len(items[1].Paths) != 0 {
		t.Fatalf("expected the first recipe to claim the folder, got %v and %v", items[0].Paths, items[1].Paths)
	}
}

func TestDynamicSiblingsSkipTheCurrentStore(t *testing.T) {
	home := testutil.FakeHome(t)
	testutil.WriteFile(t, home, "Library/pnpm/store/v3/files/a", 2*testutil.MiB)
	testutil.WriteFile(t, home, "Library/pnpm/store/v10/files/a", 2*testutil.MiB)
	current := filepath.Join(home, "Library/pnpm/store/v10")
	runner := &testutil.FakeRunner{
		Installed: map[string]bool{"pnpm": true},
		Outputs:   map[string]system.Result{"pnpm store path": {OK: true, Stdout: current + "\n"}},
	}
	old := recipe.Recipe{ID: "pnpm-old-stores", PathsFrom: &recipe.DynamicPath{Command: []string{"pnpm", "store", "path"}, Siblings: "v*"}}

	items := newScanner(home, runner, nil).Scan(context.Background(), recipe.Set{old})

	want := filepath.Join(home, "Library/pnpm/store/v3")
	if len(items[0].Paths) != 1 || items[0].Paths[0] != want {
		t.Fatalf("paths = %v, want [%s]", items[0].Paths, want)
	}
}

func TestCommandRecipes(t *testing.T) {
	home := testutil.FakeHome(t)
	docker := recipe.Recipe{ID: "docker-build", Run: []string{"docker", "builder", "prune"}, Precheck: []string{"docker", "info"}, UnavailableHint: "Docker is not running."}
	yarn := recipe.Recipe{ID: "yarn", Run: []string{"yarn", "cache", "clean"}}
	runner := &testutil.FakeRunner{Installed: map[string]bool{"docker": true}}

	items := newScanner(home, runner, nil).Scan(context.Background(), recipe.Set{docker, yarn})

	if items[0].Status != scan.StatusUnavailable || items[0].Reason != "Docker is not running." {
		t.Errorf("docker = %s %q", items[0].Status, items[0].Reason)
	}
	if items[1].Status != scan.StatusUnavailable || items[1].Reason != "yarn is not installed" {
		t.Errorf("yarn = %s %q", items[1].Status, items[1].Reason)
	}
}
