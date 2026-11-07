package engine_test

import (
	"slices"
	"testing"

	"github.com/jefferson-calmon/clean-my-cache/internal/engine"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
)

func ids(set recipe.Set) []string {
	out := make([]string, len(set))
	for i, r := range set {
		out[i] = r.ID
	}
	return out
}

func TestFilterSupportsWildcardsAndExclude(t *testing.T) {
	set := recipe.Set{{ID: "slack"}, {ID: "docker-build"}, {ID: "docker-images"}}

	only := engine.Filter(set, engine.Selection{Only: []string{"docker-*"}})
	excluded := engine.Filter(set, engine.Selection{Exclude: []string{"slack"}})

	if !slices.Equal(ids(only), []string{"docker-build", "docker-images"}) {
		t.Errorf("only = %v", ids(only))
	}
	if !slices.Equal(ids(excluded), []string{"docker-build", "docker-images"}) {
		t.Errorf("excluded = %v", ids(excluded))
	}
}

func TestCautionIsNotSelectedByDefault(t *testing.T) {
	item := scan.Item{Recipe: recipe.Recipe{ID: "user-logs", Risk: recipe.RiskCaution}, Status: scan.StatusReady}

	cases := []struct {
		selection engine.Selection
		want      bool
	}{
		{engine.Selection{}, false},
		{engine.Selection{IncludeCaution: true}, true},
		{engine.Selection{Only: []string{"user-logs"}}, true},
	}
	for _, tc := range cases {
		if got := engine.SelectedByDefault(item, tc.selection); got != tc.want {
			t.Errorf("SelectedByDefault(%+v) = %v, want %v", tc.selection, got, tc.want)
		}
	}
}

func TestOpenAppsAreNeverSelected(t *testing.T) {
	item := scan.Item{Recipe: recipe.Recipe{Risk: recipe.RiskSafe}, Status: scan.StatusAppRunning}

	if engine.SelectedByDefault(item, engine.Selection{IncludeCaution: true}) {
		t.Fatal("open app selected")
	}
}
