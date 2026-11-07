package scan

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/size"
	"github.com/jefferson-calmon/clean-my-cache/internal/system"
)

const (
	MinCleanableBytes = 1024 * 1024
	concurrency       = 8
)

type Status string

const (
	StatusReady       Status = "ready"
	StatusEmpty       Status = "empty"
	StatusAppRunning  Status = "app-running"
	StatusUnavailable Status = "unavailable"
)

type Item struct {
	Recipe   recipe.Recipe
	Status   Status
	Bytes    int64
	Measured bool
	Paths    []string
	Reason   string
}

type Scanner struct {
	Home        string
	Runner      system.Runner
	Guard       *pathguard.Guard
	RunningApps map[string]bool
}

func ResolveDynamic(ctx context.Context, runner system.Runner, dynamic *recipe.DynamicPath) []string {
	if dynamic == nil || !runner.HasCommand(dynamic.Command[0]) {
		return nil
	}

	result := runner.Run(ctx, dynamic.Command)
	printed := strings.TrimSpace(result.Stdout)
	if !result.OK || printed == "" || !filepath.IsAbs(printed) {
		return nil
	}

	printed = filepath.Clean(printed)
	if dynamic.Siblings == "" {
		return []string{printed}
	}

	var siblings []string
	for _, candidate := range pathguard.ExpandGlob(filepath.Join(filepath.Dir(printed), dynamic.Siblings)) {
		if candidate != printed {
			siblings = append(siblings, candidate)
		}
	}
	return siblings
}

func (s *Scanner) resolvePaths(ctx context.Context, r recipe.Recipe) (safe []string, rejected int) {
	candidates := pathguard.ExpandAll(r.Paths, s.Home)
	candidates = append(candidates, ResolveDynamic(ctx, s.Runner, r.PathsFrom)...)

	for _, path := range candidates {
		if s.Guard.Check(path) == "" {
			safe = append(safe, path)
			continue
		}
		rejected++
	}
	return safe, rejected
}

func (s *Scanner) measure(ctx context.Context, r recipe.Recipe) []string {
	paths := pathguard.ExpandAll(r.Measure, s.Home)
	return append(paths, ResolveDynamic(ctx, s.Runner, r.MeasureFrom)...)
}

func (s *Scanner) scanPaths(r recipe.Recipe, paths []string, rejected int) Item {
	item := Item{Recipe: r, Paths: paths, Measured: true}
	if rejected > 0 {
		item.Reason = fmt.Sprintf("%d unsafe path(s) ignored", rejected)
	}
	if len(paths) == 0 {
		item.Status = StatusEmpty
		return item
	}

	item.Bytes = size.Of(paths)
	switch {
	case item.Bytes < MinCleanableBytes:
		item.Status = StatusEmpty
	case r.App != "" && s.RunningApps[r.App]:
		item.Status = StatusAppRunning
		item.Reason = r.App + " is open"
	default:
		item.Status = StatusReady
	}
	return item
}

func (s *Scanner) scanCommand(ctx context.Context, r recipe.Recipe) Item {
	item := Item{Recipe: r}
	if !s.Runner.HasCommand(r.Bin()) {
		item.Status = StatusUnavailable
		item.Reason = r.Bin() + " is not installed"
		return item
	}

	if len(r.Precheck) > 0 && !s.Runner.Run(ctx, r.Precheck).OK {
		item.Status = StatusUnavailable
		item.Reason = r.UnavailableHint
		if item.Reason == "" {
			item.Reason = r.Bin() + " is not ready"
		}
		return item
	}

	if len(r.Measure) == 0 && r.MeasureFrom == nil {
		item.Status = StatusReady
		return item
	}

	item.Paths = s.measure(ctx, r)
	item.Bytes = size.Of(item.Paths)
	item.Measured = true
	item.Status = StatusReady
	if item.Bytes < MinCleanableBytes {
		item.Status = StatusEmpty
	}
	return item
}

type resolved struct {
	paths    []string
	rejected int
}

func parallel(n int, fn func(i int)) {
	var wg sync.WaitGroup
	slots := make(chan struct{}, concurrency)
	for i := range n {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer func() { <-slots; wg.Done() }()
			fn(i)
		}()
	}
	wg.Wait()
}

func (s *Scanner) Scan(ctx context.Context, set recipe.Set) []Item {
	resolvedPaths := make([]resolved, len(set))
	parallel(len(set), func(i int) {
		if set[i].Kind() == recipe.KindPaths {
			paths, rejected := s.resolvePaths(ctx, set[i])
			resolvedPaths[i] = resolved{paths: paths, rejected: rejected}
		}
	})

	claimed := map[string]bool{}
	for i := range resolvedPaths {
		var unique []string
		for _, path := range resolvedPaths[i].paths {
			if !claimed[path] {
				claimed[path] = true
				unique = append(unique, path)
			}
		}
		resolvedPaths[i].paths = unique
	}

	items := make([]Item, len(set))
	parallel(len(set), func(i int) {
		r := set[i]
		if r.Kind() == recipe.KindCommand {
			items[i] = s.scanCommand(ctx, r)
			return
		}
		items[i] = s.scanPaths(r, resolvedPaths[i].paths, resolvedPaths[i].rejected)
	})
	return items
}
