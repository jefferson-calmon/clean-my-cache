package recipe

import "slices"

type Category string

const (
	CategoryDev      Category = "dev"
	CategoryBrowsers Category = "browsers"
	CategoryApps     Category = "apps"
	CategorySystem   Category = "system"
)

var Categories = []Category{CategoryDev, CategoryBrowsers, CategoryApps, CategorySystem}

type Risk string

const (
	RiskSafe       Risk = "safe"
	RiskRedownload Risk = "redownload"
	RiskCaution    Risk = "caution"
)

var Risks = []Risk{RiskSafe, RiskRedownload, RiskCaution}

type Kind int

const (
	KindPaths Kind = iota
	KindCommand
)

type DynamicPath struct {
	Command  []string `yaml:"command"`
	Siblings string   `yaml:"siblings"`
}

type Recipe struct {
	ID              string       `yaml:"id"`
	Name            string       `yaml:"name"`
	Category        Category     `yaml:"category"`
	Risk            Risk         `yaml:"risk"`
	Description     string       `yaml:"description"`
	App             string       `yaml:"app"`
	Platforms       []string     `yaml:"platforms"`
	Preset          string       `yaml:"preset"`
	SupportFolder   string       `yaml:"support_folder"`
	Paths           []string     `yaml:"paths"`
	PathsFrom       *DynamicPath `yaml:"paths_from"`
	Run             []string     `yaml:"run"`
	Measure         []string     `yaml:"measure"`
	MeasureFrom     *DynamicPath `yaml:"measure_from"`
	Precheck        []string     `yaml:"precheck"`
	UnavailableHint string       `yaml:"unavailable_hint"`
	Source          string       `yaml:"-"`
}

func (r Recipe) Kind() Kind {
	if len(r.Run) > 0 {
		return KindCommand
	}
	return KindPaths
}

func (r Recipe) Bin() string {
	if len(r.Run) == 0 {
		return ""
	}
	return r.Run[0]
}

func (r Recipe) SupportsPlatform(platform string) bool {
	if len(r.Platforms) == 0 {
		return platform == "darwin"
	}
	return slices.Contains(r.Platforms, platform)
}

type Set []Recipe

func (s Set) Find(id string) (Recipe, bool) {
	for _, r := range s {
		if r.ID == id {
			return r, true
		}
	}
	return Recipe{}, false
}

func (s Set) ForPlatform(platform string) Set {
	var out Set
	for _, r := range s {
		if r.SupportsPlatform(platform) {
			out = append(out, r)
		}
	}
	return out
}

func (s Set) InCategory(category Category) Set {
	var out Set
	for _, r := range s {
		if r.Category == category {
			out = append(out, r)
		}
	}
	return out
}
