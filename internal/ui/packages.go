package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/infktd/devdash/internal/packages"
)

// PackagesView displays installed packages for a project.
type PackagesView struct {
	styles      *Styles
	packages    []packages.Package
	width       int
	height      int
	focused     bool
	titleSuffix string // Optional suffix to append to title (e.g., "[p:services]")
}

// NewPackagesView creates a new packages view.
func NewPackagesView(styles *Styles) *PackagesView {
	return &PackagesView{
		styles:   styles,
		packages: []packages.Package{},
		width:    0,
		height:   0,
	}
}

// SetPackages updates the package list.
func (pv *PackagesView) SetPackages(pkgs []packages.Package) {
	pv.packages = pkgs
}

// SetSize updates dimensions.
func (pv *PackagesView) SetSize(width, height int) {
	pv.width = width
	pv.height = height
}

// SetFocused sets focus state.
func (pv *PackagesView) SetFocused(focused bool) {
	pv.focused = focused
}

// SetTitleSuffix sets an optional suffix to append to the title.
func (pv *PackagesView) SetTitleSuffix(suffix string) {
	pv.titleSuffix = suffix
}

// View renders the packages view.
func (pv *PackagesView) View() string {
	if pv.width == 0 || pv.height == 0 {
		return ""
	}

	// Header with focus styling
	title := fmt.Sprintf("PACKAGES (%d)", len(pv.packages))
	if pv.titleSuffix != "" {
		title += " " + pv.titleSuffix
	}
	var headerStyle lipgloss.Style
	if pv.focused {
		headerStyle = lipgloss.NewStyle().
			Foreground(pv.styles.theme.Primary).
			Bold(true)
	} else {
		headerStyle = lipgloss.NewStyle().
			Foreground(pv.styles.theme.Muted)
	}
	header := headerStyle.Render(title)

	// Empty state
	if len(pv.packages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(pv.styles.theme.Muted).
			Render("No packages detected")
		content := lipgloss.NewStyle().
			Width(pv.width).
			Height(pv.height - 1).
			Align(lipgloss.Center, lipgloss.Center).
			Render(emptyMsg)

		return lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	// Column headers
	colHeaders := pv.renderColumnHeaders()

	// Package rows
	rows := pv.renderPackageRows()

	// Combine
	var lines []string
	lines = append(lines, header)
	lines = append(lines, colHeaders)
	lines = append(lines, rows...)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderColumnHeaders renders the column header row.
func (pv *PackagesView) renderColumnHeaders() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(pv.styles.theme.Muted).
		Bold(true)

	nameCol := headerStyle.Width(20).Render("NAME")
	versionCol := headerStyle.Width(15).Render("VERSION")
	typeCol := headerStyle.Width(15).Render("TYPE")

	return lipgloss.JoinHorizontal(lipgloss.Top, nameCol, versionCol, typeCol)
}

// renderPackageRows renders package data rows.
func (pv *PackagesView) renderPackageRows() []string {
	var rows []string

	// Limit to visible height (header + column headers take 2 lines)
	maxRows := pv.height - 2
	if maxRows < 0 {
		maxRows = 0
	}

	for i := 0; i < len(pv.packages) && i < maxRows; i++ {
		pkg := pv.packages[i]

		// Truncate long names
		name := pkg.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		version := pkg.Version
		if version == "" {
			version = "unknown"
		}
		if len(version) > 13 {
			version = version[:10] + "..."
		}

		typ := pkg.Type
		if len(typ) > 13 {
			typ = typ[:10] + "..."
		}

		nameCell := lipgloss.NewStyle().Width(20).Render(name)
		versionCell := lipgloss.NewStyle().Width(15).Render(version)
		typeCell := lipgloss.NewStyle().Width(15).Render(typ)

		row := lipgloss.JoinHorizontal(lipgloss.Top, nameCell, versionCell, typeCell)
		rows = append(rows, row)
	}

	return rows
}
