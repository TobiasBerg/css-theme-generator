package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/TobiasBerg/theme-generator/colors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

// ── Theme generation ─────────────────────────────────────────────────────────

type Theme struct {
	BgPrimary        string
	BgSecondary      string
	BgTertiary       string
	BgActive         string
	TextPrimary      string
	TextSecondary    string
	TextTertiary     string
	BorderPrimary    string
	BorderSecondary  string
	Success          string
	Error            string
	Warning          string
	AccentPrimary    string
	AccentSecondary  string
	AccentForeground string
	AccentMuted      string
}

func generateTheme(bg, surf, text, accent string) Theme {
	dark := !colors.IsLight(bg)

	var bgSecondary, bgTertiary, bgActive string
	if dark {
		bgSecondary = colors.Darken(bg, 5)
		bgTertiary = colors.Lighten(bg, 5)
		bgActive = colors.Lighten(bg, 10)
	} else {
		bgSecondary = colors.Lighten(bg, 5)
		bgTertiary = colors.Darken(bg, 5)
		bgActive = colors.Darken(bg, 10)
	}

	baseLightness := 60.0
	if !dark {
		baseLightness = 45
	}

	var accentFg string
	if dark {
		accentFg = colors.Darken(bg, 3)
	} else {
		accentFg = colors.Lighten(bg, 95)
	}

	return Theme{
		BgPrimary:        bg,
		BgSecondary:      bgSecondary,
		BgTertiary:       bgTertiary,
		BgActive:         bgActive,
		TextPrimary:      text,
		TextSecondary:    colors.Mix(text, bg, 0.35),
		TextTertiary:     colors.Mix(text, bg, 0.60),
		BorderPrimary:    colors.Mix(text, bg, 0.55),
		BorderSecondary:  colors.Mix(text, bg, 0.75),
		Success:          colors.HslToHex(152, 60, baseLightness),
		Error:            colors.HslToHex(0, 85, baseLightness+5),
		Warning:          colors.HslToHex(43, 95, baseLightness+5),
		AccentPrimary:    accent,
		AccentSecondary:  colors.Mix(accent, bg, 0.30),
		AccentForeground: accentFg,
		AccentMuted:      colors.Mix(accent, bg, 0.75),
	}
}

func renderCSS(name string, t Theme) string {
	return fmt.Sprintf(`[data-theme="%s"] {
    --bg-primary: %s;
    --bg-secondary: %s;
    --bg-tertiary: %s;
    --bg-active: %s;
    --text-primary: %s;
    --text-secondary: %s;
    --text-tertiary: %s;
    --border-primary: %s;
    --border-secondary: %s;
    --success: %s;
    --error: %s;
    --warning: %s;
    --accent-primary: %s;
    --accent-secondary: %s;
    --accent-foreground: %s;
    --accent-muted: %s;
}`,
		name,
		t.BgPrimary, t.BgSecondary, t.BgTertiary, t.BgActive,
		t.TextPrimary, t.TextSecondary, t.TextTertiary,
		t.BorderPrimary, t.BorderSecondary,
		t.Success, t.Error, t.Warning,
		t.AccentPrimary, t.AccentSecondary, t.AccentForeground, t.AccentMuted,
	)
}

// ── Terminal preview ──────────────────────────────────────────────────────────

func swatchBlock(hex string) string {
	c, _ := colors.HexToRGB(hex)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm  \x1b[0m", int(c.R), int(c.G), int(c.B))
}

func printPreview(t Theme) {
	dim := color.New(color.Faint)
	bold := color.New(color.Bold)

	fmt.Println()
	bold.Println("  Preview")
	fmt.Println()

	rows := []struct{ label, hex string }{
		{"bg-primary      ", t.BgPrimary},
		{"bg-secondary    ", t.BgSecondary},
		{"bg-tertiary     ", t.BgTertiary},
		{"bg-active       ", t.BgActive},
		{"text-primary    ", t.TextPrimary},
		{"text-secondary  ", t.TextSecondary},
		{"text-tertiary   ", t.TextTertiary},
		{"border-primary  ", t.BorderPrimary},
		{"border-secondary", t.BorderSecondary},
		{"success         ", t.Success},
		{"error           ", t.Error},
		{"warning         ", t.Warning},
		{"accent-primary  ", t.AccentPrimary},
		{"accent-secondary", t.AccentSecondary},
		{"accent-foreground", t.AccentForeground},
		{"accent-muted    ", t.AccentMuted},
	}

	for _, row := range rows {
		dim.Printf("  --%-18s", row.label)
		fmt.Printf(" %s  %s\n", row.hex, swatchBlock(row.hex))
	}
	fmt.Println()
}

// ── Input helpers ─────────────────────────────────────────────────────────────

var hexRe = regexp.MustCompile(`(?i)^#?[0-9a-f]{3}([0-9a-f]{3})?$`)

func normalizeHex(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}
	if !hexRe.MatchString(s) {
		return "", false
	}
	return strings.ToLower(s), true
}

func prompt(scanner *bufio.Scanner, label string) string {
	bold := color.New(color.Bold)
	for {
		bold.Printf("  %s: ", label)
		scanner.Scan()
		val := strings.TrimSpace(scanner.Text())
		if h, ok := normalizeHex(val); ok {
			return h
		}
		color.New(color.FgRed).Printf("  ✗ %q is not a valid hex color — try again\n", val)
	}
}

func autoDetect(clrs []string) (bg, surf, text, accent string) {
	sorted := make([]string, len(clrs))
	copy(sorted, clrs)
	// bubble sort by luminance ascending
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if colors.Luminance(sorted[i]) > colors.Luminance(sorted[j]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted[0], sorted[1], sorted[3], sorted[2]
}

// ── Main ──────────────────────────────────────────────────────────────────────

func CreateTheneCMD() func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		scanner := bufio.NewScanner(os.Stdin)

		header := color.New(color.Bold, color.FgCyan)
		dim := color.New(color.Faint)
		green := color.New(color.FgGreen)

		fmt.Println()
		header.Println("  ◆ themegen — CSS theme generator")
		dim.Println("  Paste 4 hex colors from colorhunt.co")
		fmt.Println()

		colors := make([]string, 4)
		labels := []string{
			"Color 1 (background base)",
			"Color 2 (surface / card)",
			"Color 3 (text / foreground)",
			"Color 4 (accent)",
		}

		for i, label := range labels {
			colors[i] = prompt(scanner, label)
		}

		fmt.Println()
		dim.Println("  Auto-detect roles? Reorders by luminance (darkest→bg, lightest→text).")
		color.New(color.Bold).Print("  Use auto-detect? [y/N]: ")
		scanner.Scan()
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))

		bg, surf, text, accent := colors[0], colors[1], colors[2], colors[3]
		if answer == "y" || answer == "yes" {
			bg, surf, text, accent = autoDetect(colors)
			dim.Printf("  → bg=%s  surf=%s  text=%s  accent=%s\n", bg, surf, text, accent)
		}

		fmt.Println()
		color.New(color.Bold).Print("  Theme name [my-theme]: ")
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			name = "my-theme"
		}

		theme := generateTheme(bg, surf, text, accent)
		css := renderCSS(name, theme)

		printPreview(theme)

		// Output CSS
		green.Println("  ── Generated CSS ─────────────────────────────────")
		fmt.Println()
		fmt.Println(css)
		fmt.Println()

		// Optionally save to file
		color.New(color.Bold).Print("  Save to file? (leave blank to skip): ")
		scanner.Scan()
		outPath := strings.TrimSpace(scanner.Text())
		if outPath != "" {
			// Append if file exists
			f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				color.New(color.FgRed).Printf("  ✗ Could not open %s: %v\n", outPath, err)
				os.Exit(1)
			}
			defer f.Close()
			fmt.Fprintln(f, "")
			fmt.Fprintln(f, css)
			green.Printf("  ✓ Appended to %s\n\n", outPath)
		} else {
			dim.Println("  (copy the CSS above into your stylesheet)")
		}

		return nil
	}
}
