package card

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultTheme    = "default"
	LightTheme      = "light"
	DraculaTheme    = "dracula"
	TokyoNightTheme = "tokyonight"
	GruvboxTheme    = "gruvbox"
)

var ErrUnknownTheme = errors.New("unknown card theme")

type Theme struct {
	Name       string
	Background string
	Border     string
	Title      string
	Text       string
	Value      string
	Accent     string
	Track      string
}

var themes = map[string]Theme{
	DefaultTheme: {
		Name:       DefaultTheme,
		Background: "#0d1117",
		Border:     "#30363d",
		Title:      "#f0f6fc",
		Text:       "#7d8590",
		Value:      "#e6edf3",
		Accent:     "#3fb950",
		Track:      "#21262d",
	},
	LightTheme: {
		Name:       LightTheme,
		Background: "#ffffff",
		Border:     "#d0d7de",
		Title:      "#1f2328",
		Text:       "#656d76",
		Value:      "#1f2328",
		Accent:     "#0969da",
		Track:      "#eaeef2",
	},
	DraculaTheme: {
		Name:       DraculaTheme,
		Background: "#282a36",
		Border:     "#44475a",
		Title:      "#f8f8f2",
		Text:       "#6272a4",
		Value:      "#f8f8f2",
		Accent:     "#bd93f9",
		Track:      "#343746",
	},
	TokyoNightTheme: {
		Name:       TokyoNightTheme,
		Background: "#1a1b26",
		Border:     "#2f334d",
		Title:      "#c0caf5",
		Text:       "#565f89",
		Value:      "#c0caf5",
		Accent:     "#7aa2f7",
		Track:      "#24283b",
	},
	GruvboxTheme: {
		Name:       GruvboxTheme,
		Background: "#282828",
		Border:     "#504945",
		Title:      "#ebdbb2",
		Text:       "#928374",
		Value:      "#ebdbb2",
		Accent:     "#fabd2f",
		Track:      "#3c3836",
	},
}

func ResolveTheme(name string) (Theme, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		normalized = DefaultTheme
	}

	theme, found := themes[normalized]
	if !found {
		return Theme{}, fmt.Errorf(
			"%w: %q",
			ErrUnknownTheme,
			normalized,
		)
	}
	return theme, nil
}
