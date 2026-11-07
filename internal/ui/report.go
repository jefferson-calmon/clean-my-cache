package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/jefferson-calmon/clean-my-cache/internal/clean"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
)

const (
	nameWidth = 34
	sizeWidth = 9
)

func SizeOf(item scan.Item) string {
	if !item.Measured {
		return "?"
	}
	return FormatBytes(item.Bytes)
}

func describeStatus(item scan.Item) string {
	switch item.Status {
	case scan.StatusReady:
		return RiskLabel(item.Recipe.Risk)
	case scan.StatusEmpty:
		return Dim.Render("nothing worth cleaning")
	case scan.StatusAppRunning:
		return Yellow.Render(fmt.Sprintf("skipped, %s is open", item.Recipe.App))
	default:
		return Dim.Render(item.Reason)
	}
}

func row(item scan.Item) string {
	size := ""
	if item.Status != scan.StatusUnavailable {
		size = SizeOf(item)
	}
	return fmt.Sprintf("  %s %-*s%s  %s",
		StatusSymbol[item.Status], nameWidth, item.Recipe.Name,
		Bold.Render(fmt.Sprintf("%*s", sizeWidth, size)), describeStatus(item))
}

func BySizeDesc(items []scan.Item) []scan.Item {
	out := slices.Clone(items)
	slices.SortStableFunc(out, func(a, b scan.Item) int {
		return cmp.Compare(b.Bytes, a.Bytes)
	})
	return out
}

func Report(items []scan.Item, verbose bool) string {
	var sections []string
	for _, category := range recipe.Categories {
		var rows []string
		for _, item := range BySizeDesc(items) {
			if item.Recipe.Category != category {
				continue
			}
			if !verbose && item.Status != scan.StatusReady && item.Status != scan.StatusAppRunning {
				continue
			}
			rows = append(rows, row(item))
		}
		if len(rows) > 0 {
			sections = append(sections, Title.Render(CategoryTitle[category])+"\n"+strings.Join(rows, "\n"))
		}
	}

	if len(sections) == 0 {
		return Dim.Render("  Nothing to clean. Your caches are already tidy.")
	}
	return strings.Join(sections, "\n\n")
}

func Total(items []scan.Item) int64 {
	var total int64
	for _, item := range items {
		total += item.Bytes
	}
	return total
}

func WithStatus(items []scan.Item, status scan.Status) []scan.Item {
	var out []scan.Item
	for _, item := range items {
		if item.Status == status {
			out = append(out, item)
		}
	}
	return out
}

func OutcomeLine(o clean.Outcome) string {
	if !o.OK {
		return fmt.Sprintf("  %s %s: %s", SymbolFail, o.Item.Recipe.Name, Red.Render(o.Err.Error()))
	}
	freed := Dim.Render("done")
	if o.FreedBytes > 0 {
		freed = Bold.Render(FormatBytes(o.FreedBytes))
	}
	return fmt.Sprintf("  %s %-*s%s", SymbolOK, nameWidth, o.Item.Recipe.Name, freed)
}

func SkippedApps(items []scan.Item) string {
	running := WithStatus(items, scan.StatusAppRunning)
	if len(running) == 0 {
		return ""
	}

	apps := make([]string, len(running))
	ids := make([]string, len(running))
	for i, item := range running {
		apps[i] = item.Recipe.App
		ids[i] = item.Recipe.ID
	}

	return fmt.Sprintf("%s %s skipped because these apps are open: %s.\n%s",
		SymbolWarn, FormatBytes(Total(running)), strings.Join(apps, ", "),
		Dim.Render("  Quit them and run: clean-my-cache --only "+strings.Join(ids, ",")))
}
