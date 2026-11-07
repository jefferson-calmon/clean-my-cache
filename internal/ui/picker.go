package ui

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
)

var ErrCancelled = errors.New("cancelled")

const pickerHeight = 18

func optionLabel(item scan.Item) string {
	return fmt.Sprintf("%-32s %9s  %s  %s",
		item.Recipe.Name, SizeOf(item), RiskLabel(item.Recipe.Risk), Dim.Render(item.Recipe.Description))
}

func Pick(ready []scan.Item, preselected func(scan.Item) bool, trash bool) ([]scan.Item, error) {
	sorted := BySizeDesc(ready)
	options := make([]huh.Option[string], len(sorted))
	for i, item := range sorted {
		options[i] = huh.NewOption(optionLabel(item), item.Recipe.ID).Selected(preselected(item))
	}

	var chosen []string
	pick := huh.NewMultiSelect[string]().
		Title("What should be cleaned?").
		Options(options...).
		Height(pickerHeight).
		Value(&chosen)

	if err := huh.NewForm(huh.NewGroup(pick)).Run(); err != nil {
		return nil, ErrCancelled
	}

	selected := make([]scan.Item, 0, len(chosen))
	for _, item := range sorted {
		if slices.Contains(chosen, item.Recipe.ID) {
			selected = append(selected, item)
		}
	}
	if len(selected) == 0 {
		return nil, nil
	}

	action := "Delete"
	if trash {
		action = "Move to Trash"
	}

	confirmed := false
	confirm := huh.NewConfirm().
		Title(fmt.Sprintf("%s %d item(s), about %s?", action, len(selected), FormatBytes(Total(selected)))).
		Affirmative("Yes, clean").
		Negative("No").
		Value(&confirmed)

	if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil || !confirmed {
		return nil, ErrCancelled
	}
	return selected, nil
}
