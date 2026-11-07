package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
)

var (
	Bold   = lipgloss.NewStyle().Bold(true)
	Dim    = lipgloss.NewStyle().Faint(true)
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Title  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
)

var (
	SymbolOK   = Green.Render("+")
	SymbolSkip = Dim.Render("-")
	SymbolWarn = Yellow.Render("!")
	SymbolInfo = Dim.Render("~")
	SymbolFail = Red.Render("x")
)

var StatusSymbol = map[scan.Status]string{
	scan.StatusReady:       SymbolOK,
	scan.StatusEmpty:       SymbolSkip,
	scan.StatusAppRunning:  SymbolWarn,
	scan.StatusUnavailable: SymbolInfo,
}

var riskText = map[recipe.Risk]string{
	recipe.RiskSafe:       "safe",
	recipe.RiskRedownload: "re-download",
	recipe.RiskCaution:    "caution",
}

var riskStyle = map[recipe.Risk]lipgloss.Style{
	recipe.RiskSafe:       Green,
	recipe.RiskRedownload: Yellow,
	recipe.RiskCaution:    Red,
}

var RiskMeaning = map[recipe.Risk]string{
	recipe.RiskSafe:       "Nothing is lost. The app rebuilds it on its own.",
	recipe.RiskRedownload: "Nothing is lost, but it is downloaded or rebuilt next time, which can take a while.",
	recipe.RiskCaution:    "Usually fine, but you might want to keep it. Never selected unless you ask for it.",
}

var CategoryTitle = map[recipe.Category]string{
	recipe.CategoryDev:      "Developer tools",
	recipe.CategoryBrowsers: "Browsers",
	recipe.CategoryApps:     "Apps",
	recipe.CategorySystem:   "System",
}

func RiskLabel(risk recipe.Risk) string {
	return riskStyle[risk].Render(riskText[risk])
}

func RiskText(risk recipe.Risk) string {
	return riskText[risk]
}

var units = []string{"B", "KB", "MB", "GB", "TB"}

func FormatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}

	if value >= 100 || unit < 2 {
		return fmt.Sprintf("%.0f %s", value, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}
