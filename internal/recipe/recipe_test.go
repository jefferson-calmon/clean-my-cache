package recipe_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/recipes"
)

func TestBuiltinRecipesAreValid(t *testing.T) {
	set, err := recipe.Load(recipes.Files, "recipes")
	if err != nil {
		t.Fatal(err)
	}
	if len(set) < 38 {
		t.Fatalf("expected at least 38 recipes, got %d", len(set))
	}
	if _, err := recipe.Merge(set); err != nil {
		t.Fatal(err)
	}
}

func TestBuiltinRecipesLiveInFolderOfTheirCategory(t *testing.T) {
	set, err := recipe.Load(recipes.Files, "recipes")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range set {
		want := "recipes/" + string(r.Category) + "/" + r.ID + ".yaml"
		if r.Source != want {
			t.Errorf("%s should be at %s, found at %s", r.ID, want, r.Source)
		}
	}
}

func TestElectronPresetExpandsCacheFolders(t *testing.T) {
	set, err := recipe.Load(recipes.Files, "recipes")
	if err != nil {
		t.Fatal(err)
	}
	slack, ok := set.Find("slack")
	if !ok {
		t.Fatal("slack recipe missing")
	}
	if slack.Paths[0] != "~/Library/Application Support/Slack/Cache" {
		t.Fatalf("unexpected first path %q", slack.Paths[0])
	}
}

const validBody = `
name: Tool
category: dev
risk: safe
description: A cache that is rebuilt automatically.
`

func TestValidationRejectsUnsafeRecipes(t *testing.T) {
	cases := map[string]struct {
		yaml string
		want string
	}{
		"shell command":     {"id: bad\nrun: [bash, -c, rm -rf ~]" + validBody, `cannot run "bash"`},
		"absolute bin":      {"id: bad\nrun: [/bin/rm, -rf]" + validBody, "must name a program"},
		"outside roots":     {"id: bad\npaths: [~/Documents]" + validBody, "allowed cache folder"},
		"dot dot":           {"id: bad\npaths: [~/Library/Caches/../../Documents]" + validBody, "'..'"},
		"both kinds":        {"id: bad\npaths: [~/Library/Caches/x]\nrun: [npm, cache]" + validBody, "not both"},
		"neither kind":      {"id: bad" + validBody, "not both"},
		"unknown field":     {"id: bad\npaths: [~/Library/Caches/x]\ndelete_everything: true" + validBody, "delete_everything"},
		"bad id":            {"id: Bad_ID\npaths: [~/Library/Caches/x]" + validBody, "kebab-case"},
		"short description": {"id: bad\nname: x\ncategory: dev\nrisk: safe\ndescription: short\npaths: [~/Library/Caches/x]\n", "description"},
		"bad risk":          {"id: bad\nname: x\ncategory: dev\nrisk: yolo\ndescription: A cache that is rebuilt automatically.\npaths: [~/Library/Caches/x]\n", "risk"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := recipe.Load(fstest.MapFS{"dev/bad.yaml": {Data: []byte(tc.yaml)}}, "test")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestMergeRejectsDuplicateIDs(t *testing.T) {
	r := recipe.Recipe{ID: "npm", Source: "a"}
	other := recipe.Recipe{ID: "npm", Source: "b"}

	if _, err := recipe.Merge(recipe.Set{r}, recipe.Set{other}); err == nil {
		t.Fatal("expected duplicate id error")
	}
}
