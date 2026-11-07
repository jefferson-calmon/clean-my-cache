package recipe

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
)

var idPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var forbiddenBins = []string{"sh", "bash", "zsh", "fish", "dash", "ksh", "env", "sudo", "rm", "xargs", "eval"}

const minDescriptionLength = 20

func checkCommand(field string, command []string) error {
	if len(command) == 0 {
		return nil
	}
	bin := command[0]
	if strings.Contains(bin, "/") {
		return fmt.Errorf("%s must name a program, not a path: %q", field, bin)
	}
	if slices.Contains(forbiddenBins, bin) {
		return fmt.Errorf("%s cannot run %q", field, bin)
	}
	return nil
}

func checkPaths(field string, paths []string, deletable bool) []error {
	var errs []error
	for _, p := range paths {
		switch {
		case strings.Contains(p, ".."):
			errs = append(errs, fmt.Errorf("%s cannot contain '..': %q", field, p))
		case !strings.HasPrefix(p, "~/"):
			errs = append(errs, fmt.Errorf("%s must start with ~/: %q", field, p))
		case deletable && !pathguard.IsUnderAllowedRoot(p):
			errs = append(errs, fmt.Errorf("%s must live inside an allowed cache folder: %q", field, p))
		}
	}
	return errs
}

func Validate(r Recipe) error {
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if !idPattern.MatchString(r.ID) {
		add(fmt.Errorf("id must be lowercase kebab-case: %q", r.ID))
	}
	if strings.TrimSpace(r.Name) == "" {
		add(errors.New("name is required"))
	}
	if !slices.Contains(Categories, r.Category) {
		add(fmt.Errorf("category must be one of %v: %q", Categories, r.Category))
	}
	if !slices.Contains(Risks, r.Risk) {
		add(fmt.Errorf("risk must be one of %v: %q", Risks, r.Risk))
	}
	if len(strings.TrimSpace(r.Description)) < minDescriptionLength {
		add(fmt.Errorf("description must explain the item in at least %d characters", minDescriptionLength))
	}
	if r.Preset != "" && r.Preset != "electron" {
		add(fmt.Errorf("unknown preset %q", r.Preset))
	}
	if r.Preset == "electron" && r.SupportFolder == "" {
		add(errors.New("preset electron needs support_folder"))
	}

	hasPaths := len(r.Paths) > 0 || r.PathsFrom != nil
	hasRun := len(r.Run) > 0
	if hasPaths == hasRun {
		add(errors.New("a recipe needs either paths (paths, preset or paths_from) or run, not both"))
	}

	errs = append(errs, checkPaths("paths", r.Paths, true)...)
	errs = append(errs, checkPaths("measure", r.Measure, false)...)
	add(checkCommand("run", r.Run))
	add(checkCommand("precheck", r.Precheck))
	if r.PathsFrom != nil {
		add(checkCommand("paths_from.command", r.PathsFrom.Command))
	}
	if r.MeasureFrom != nil {
		add(checkCommand("measure_from.command", r.MeasureFrom.Command))
	}

	return errors.Join(errs...)
}
