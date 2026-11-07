package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/jefferson-calmon/clean-my-cache/internal/clean"
	"github.com/jefferson-calmon/clean-my-cache/internal/engine"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
	"github.com/jefferson-calmon/clean-my-cache/internal/ui"
)

const cleanConcurrency = 4

type jsonItem struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Risk     string   `json:"risk"`
	Status   string   `json:"status"`
	Bytes    *int64   `json:"bytes"`
	Paths    []string `json:"paths"`
	Reason   string   `json:"reason,omitempty"`
}

type jsonOutcome struct {
	ID         string `json:"id"`
	OK         bool   `json:"ok"`
	FreedBytes int64  `json:"freedBytes"`
	Error      string `json:"error,omitempty"`
}

func toJSONItems(items []scan.Item) []jsonItem {
	out := make([]jsonItem, len(items))
	for i, item := range items {
		var bytes *int64
		if item.Measured {
			bytes = &item.Bytes
		}
		out[i] = jsonItem{
			ID: item.Recipe.ID, Name: item.Recipe.Name, Category: string(item.Recipe.Category),
			Risk: string(item.Recipe.Risk), Status: string(item.Status), Bytes: bytes,
			Paths: append([]string{}, item.Paths...), Reason: item.Reason,
		}
	}
	return out
}

func (a *app) writeJSON(value any) error {
	encoder := json.NewEncoder(a.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (a *app) choose(ready []scan.Item, opts options) ([]scan.Item, error) {
	preselected := func(item scan.Item) bool { return engine.SelectedByDefault(item, opts.selection()) }
	if a.interactive && !opts.yes {
		return ui.Pick(ready, preselected, opts.trash)
	}

	var selected []scan.Item
	for _, item := range ready {
		if preselected(item) {
			selected = append(selected, item)
		}
	}
	return selected, nil
}

func (a *app) cleanAll(ctx context.Context, cleaner *clean.Cleaner, items []scan.Item, printLines bool) []clean.Outcome {
	outcomes := make([]clean.Outcome, len(items))
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, cleanConcurrency)

	for i, item := range items {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer func() { <-slots; wg.Done() }()
			outcome := cleaner.Clean(ctx, item)
			outcomes[i] = outcome
			if printLines {
				mu.Lock()
				a.println(ui.OutcomeLine(outcome))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return outcomes
}

func (a *app) summarize(outcomes []clean.Outcome, items []scan.Item, trash bool) {
	var freed int64
	failed := 0
	for _, o := range outcomes {
		freed += o.FreedBytes
		if !o.OK {
			failed++
		}
	}

	suffix := ""
	if trash {
		suffix = ui.Dim.Render(" (moved to Trash)")
	}

	a.println()
	a.println(ui.SymbolOK, ui.Green.Bold(true).Render("Freed "+ui.FormatBytes(freed))+suffix)
	if failed > 0 {
		a.println(ui.SymbolFail, fmt.Sprintf("%d item(s) failed.", failed))
	}
	if skipped := ui.SkippedApps(items); skipped != "" {
		a.println(skipped)
	}
}

func (a *app) runClean(ctx context.Context, opts options) error {
	if err := requireMacOS(); err != nil {
		return err
	}

	e, err := a.engine(ctx)
	if err != nil {
		return err
	}

	items := ui.WithSpinner(a.errOut, a.interactive && !opts.json, "Looking for caches", func() []scan.Item {
		return e.Scan(ctx, opts.selection())
	})
	ready := ui.WithStatus(items, scan.StatusReady)

	if opts.json && opts.dryRun {
		return a.writeJSON(toJSONItems(items))
	}

	if !opts.json {
		a.println()
		a.println(ui.Report(items, opts.verbose))
		a.println()
	}

	if opts.dryRun {
		a.println(ui.Bold.Render(ui.FormatBytes(ui.Total(ready))), "can be cleaned.", ui.Dim.Render("Nothing was deleted (--dry-run)."))
		if skipped := ui.SkippedApps(items); skipped != "" {
			a.println(skipped)
		}
		return nil
	}

	if len(ready) == 0 {
		return nil
	}

	if !a.interactive && !opts.yes {
		a.println("Nothing was deleted. Add", ui.Bold.Render("--yes"), "to clean without questions.")
		return errExitCode
	}

	selected, err := a.choose(ready, opts)
	if errors.Is(err, ui.ErrCancelled) {
		a.println()
		a.println(ui.Dim.Render("Nothing was deleted."))
		return errExitCode
	}
	if len(selected) == 0 {
		a.println(ui.Dim.Render("Nothing selected. Nothing was deleted."))
		return nil
	}

	mode := clean.ModeDelete
	if opts.trash {
		mode = clean.ModeTrash
	}
	outcomes := a.cleanAll(ctx, e.Cleaner(mode), selected, !opts.json)

	failed := false
	for _, o := range outcomes {
		failed = failed || !o.OK
	}

	if opts.json {
		results := make([]jsonOutcome, len(outcomes))
		for i, o := range outcomes {
			results[i] = jsonOutcome{ID: o.Item.Recipe.ID, OK: o.OK, FreedBytes: o.FreedBytes}
			if o.Err != nil {
				results[i].Error = o.Err.Error()
			}
		}
		if err := a.writeJSON(results); err != nil {
			return err
		}
	} else {
		a.summarize(outcomes, items, opts.trash)
	}

	if failed {
		return errExitCode
	}
	return nil
}
