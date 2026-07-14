package card

import (
	"errors"
	"testing"
)

func TestResolveTheme(t *testing.T) {
	defaultTheme := Theme{
		Name:       DefaultTheme,
		Background: "#0d1117",
		Border:     "#30363d",
		Title:      "#f0f6fc",
		Text:       "#7d8590",
		Value:      "#e6edf3",
		Accent:     "#3fb950",
		Track:      "#21262d",
	}

	lightTheme := Theme{
		Name:       LightTheme,
		Background: "#ffffff",
		Border:     "#d0d7de",
		Title:      "#1f2328",
		Text:       "#656d76",
		Value:      "#1f2328",
		Accent:     "#0969da",
		Track:      "#eaeef2",
	}

	tests := []struct {
		name  string
		input string
		want  Theme
	}{
		{
			name:  "empty name uses default",
			input: "",
			want:  defaultTheme,
		},
		{
			name:  "whitespace uses default",
			input: "   ",
			want:  defaultTheme,
		},
		{
			name:  "explicit default",
			input: "default",
			want:  defaultTheme,
		},
		{
			name:  "normalizes light theme",
			input: "  LIGHT  ",
			want:  lightTheme,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveTheme(test.input)
			if err != nil {
				t.Fatalf("resolve theme: %v", err)
			}

			if got != test.want {
				t.Fatalf(
					"unexpected theme: got %+v, want %+v",
					got,
					test.want,
				)
			}
		})
	}
}

func TestResolveThemeRejectsUnknownTheme(t *testing.T) {
	theme, err := ResolveTheme("unknown")

	if !errors.Is(err, ErrUnknownTheme) {
		t.Fatalf("expected ErrUnknownTheme, got %v", err)
	}

	if theme != (Theme{}) {
		t.Fatalf("expected zero theme, got %+v", theme)
	}
}
