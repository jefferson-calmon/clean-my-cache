package cli

import (
	"fmt"
	"strings"

	"github.com/jefferson-calmon/clean-my-cache/internal/engine"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/ui"
)

type jsonRecipe struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Risk        string `json:"risk"`
	Description string `json:"description"`
}

func (a *app) runList(asJSON bool) error {
	set, err := engine.LoadRecipes(a.home)
	if err != nil {
		return fmt.Errorf("invalid recipe: %w", err)
	}

	if asJSON {
		out := make([]jsonRecipe, len(set))
		for i, r := range set {
			out[i] = jsonRecipe{ID: r.ID, Name: r.Name, Category: string(r.Category), Risk: string(r.Risk), Description: r.Description}
		}
		return a.writeJSON(out)
	}

	for _, category := range recipe.Categories {
		a.println(ui.Title.Render(ui.CategoryTitle[category]))
		for _, r := range set.InCategory(category) {
			a.println(fmt.Sprintf("  %-18s%-34s%s", r.ID, r.Name, ui.RiskLabel(r.Risk)))
		}
		a.println()
	}
	a.println(ui.Dim.Render("See what an item does: clean-my-cache explain <id>"))
	a.println(ui.Dim.Render("Add your own recipes in " + engine.UserRecipesDir(a.home)))
	return nil
}

func label(name string) string {
	return fmt.Sprintf("%-10s", name+":")
}

func (a *app) runExplain(id string) error {
	set, err := engine.LoadRecipes(a.home)
	if err != nil {
		return fmt.Errorf("invalid recipe: %w", err)
	}

	r, ok := set.Find(id)
	if !ok {
		return fmt.Errorf("unknown id %q. Run clean-my-cache list to see every id", id)
	}

	a.println(ui.Bold.Render(r.Name))
	a.println(r.Description)
	a.println()
	a.println(label("Risk"), ui.RiskLabel(r.Risk), ui.Dim.Render(ui.RiskMeaning[r.Risk]))
	if r.App != "" {
		a.println(label("Skipped"), "while", r.App, "is open")
	}

	if r.Kind() == recipe.KindCommand {
		a.println(label("Runs"), strings.Join(r.Run, " "))
	}
	for i, path := range r.Paths {
		name := ""
		if i == 0 {
			name = "Folders:"
		}
		a.println(fmt.Sprintf("%-10s", name), path)
	}
	if r.PathsFrom != nil {
		a.println(label("Folders"), "found by running", strings.Join(r.PathsFrom.Command, " "))
	}
	a.println(label("Recipe"), ui.Dim.Render(r.Source))
	return nil
}
