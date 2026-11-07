package engine

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jefferson-calmon/clean-my-cache/internal/clean"
	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
	"github.com/jefferson-calmon/clean-my-cache/internal/system"
	"github.com/jefferson-calmon/clean-my-cache/recipes"
)

type Selection struct {
	Only           []string
	Exclude        []string
	IncludeCaution bool
}

func matchesAny(id string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, id); matched {
			return true
		}
	}
	return false
}

func Filter(set recipe.Set, selection Selection) recipe.Set {
	var out recipe.Set
	for _, r := range set {
		if len(selection.Only) > 0 && !matchesAny(r.ID, selection.Only) {
			continue
		}
		if matchesAny(r.ID, selection.Exclude) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func SelectedByDefault(item scan.Item, selection Selection) bool {
	if item.Status != scan.StatusReady {
		return false
	}
	if len(selection.Only) > 0 || selection.IncludeCaution {
		return true
	}
	return item.Recipe.Risk != recipe.RiskCaution
}

func UserRecipesDir(home string) string {
	return filepath.Join(home, ".config", "clean-my-cache", "recipes")
}

func LoadRecipes(home string) (recipe.Set, error) {
	builtin, err := recipe.Load(recipes.Files, "recipes")
	if err != nil {
		return nil, err
	}
	custom, err := recipe.LoadDir(UserRecipesDir(home))
	if err != nil {
		return nil, err
	}
	return recipe.Merge(builtin, custom)
}

type Engine struct {
	Home    string
	Recipes recipe.Set
	runner  system.Runner
	guard   *pathguard.Guard
	scanner *scan.Scanner
}

func New(ctx context.Context, home string, runner system.Runner, set recipe.Set) *Engine {
	guard := pathguard.New(home)
	return &Engine{
		Home:    home,
		Recipes: set,
		runner:  runner,
		guard:   guard,
		scanner: &scan.Scanner{
			Home:        home,
			Runner:      runner,
			Guard:       guard,
			RunningApps: system.RunningApps(ctx, runner),
		},
	}
}

func (e *Engine) Scan(ctx context.Context, selection Selection) []scan.Item {
	return e.scanner.Scan(ctx, Filter(e.Recipes.ForPlatform(runtime.GOOS), selection))
}

func (e *Engine) Cleaner(mode clean.Mode) *clean.Cleaner {
	return &clean.Cleaner{Home: e.Home, Runner: e.runner, Guard: e.guard, Mode: mode, Now: time.Now}
}
