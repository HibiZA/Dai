package style

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Text styling functions
var (
	// Headers and titles
	Title     = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	Subtitle  = color.New(color.FgCyan).SprintFunc()
	Header    = color.New(color.FgBlue, color.Bold).SprintFunc()
	Subheader = color.New(color.FgHiBlue).SprintFunc()

	// Highlights
	Highlight = color.New(color.FgHiYellow).SprintFunc()
	Info      = color.New(color.FgHiWhite).SprintFunc()
	Success   = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	Warning   = color.New(color.FgHiYellow, color.Bold).SprintFunc()
	Error     = color.New(color.FgHiRed, color.Bold).SprintFunc()

	// Package and version formatting
	Package = color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	Version = color.New(color.FgHiGreen).SprintFunc()

	// Severity levels
	Critical = color.New(color.BgRed, color.FgHiWhite, color.Bold).SprintFunc()
	High     = color.New(color.FgHiRed, color.Bold).SprintFunc()
	Medium   = color.New(color.FgHiYellow, color.Bold).SprintFunc()
	Low      = color.New(color.FgHiBlue, color.Bold).SprintFunc()

	// Other elements
	Bullet = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	URL    = color.New(color.FgHiBlue, color.Underline).SprintFunc()
)

// GetSeverityColor returns a colored version of the severity string
func GetSeverityColor(severity string) string {
	sev := strings.ToUpper(severity)
	switch sev {
	case "CRITICAL":
		return Critical(" " + sev + " ")
	case "HIGH":
		return High(sev)
	case "MEDIUM":
		return Medium(sev)
	case "LOW":
		return Low(sev)
	default:
		return severity
	}
}

// FormatPackage returns a styled package@version string
func FormatPackage(name, version string) string {
	return fmt.Sprintf("%s@%s", Package(name), Version(version))
}

// BoldItalicize applies bold and italic effects to text
func BoldItalicize(text string) string {
	return color.New(color.Bold, color.Italic).Sprint(text)
}

// Banner returns a colorful ASCII art banner for the CLI
func Banner() string {
	lines := []string{
		"   ___       _   ",
		"  / _ \\__ _ (_)",
		" / // / _ `/ / ",
		"/____/\\_,_/_/  ",
		"                ",
	}

	var coloredBanner strings.Builder
	color1 := color.New(color.FgHiMagenta, color.Bold)
	color2 := color.New(color.FgHiCyan, color.Bold)

	for i, line := range lines {
		if i%2 == 0 {
			coloredBanner.WriteString(color1.Sprint(line) + "\n")
		} else {
			coloredBanner.WriteString(color2.Sprint(line) + "\n")
		}
	}

	tagline := color.New(color.FgHiGreen, color.Italic).Sprint("AI‑Backed Dependency Upgrade Advisor")
	coloredBanner.WriteString(tagline)

	return coloredBanner.String()
}

// Divider returns a colorful divider line
func Divider() string {
	divider := strings.Repeat("─", 50)
	return color.New(color.FgHiBlue).Sprint(divider)
}

// FormatBulletList returns a formatted bullet list with colorful bullets
func FormatBulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var result strings.Builder
	bulletChar := Bullet("•")

	for _, item := range items {
		result.WriteString(fmt.Sprintf(" %s %s\n", bulletChar, item))
	}

	return result.String()
}
