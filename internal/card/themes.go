package card

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultTheme = "default"
	LightTheme   = "light"
)

var ErrUnknownTheme = errors.New("unknown card theme")

type Theme struct {
	Name       string
	Background string
	Border     string
	Title      string
	Text       string
	Accent     string
}

var themes = map[string]Theme{
	DefaultTheme: {
		Name:       DefaultTheme,
		Background: "#0d1117",
		Border:     "#30363d",
		Title:      "#f0f6fc",
		Text:       "#8b949e",
		Accent:     "#2f81f7",
	},
	LightTheme: {
		Name:       LightTheme,
		Background: "#ffffff",
		Border:     "#d0d7de",
		Title:      "#1f2328",
		Text:       "#656d76",
		Accent:     "#0969da",
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
