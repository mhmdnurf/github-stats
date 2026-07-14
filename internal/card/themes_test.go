package card

import (
	"errors"
	"testing"
)

func TestResolveTheme(t *testing.T) {
	defaultTheme := Theme{
		Name:       "default",
		Background: "#0d1117",
		Border:     "#30363d",
		Title:      "#f0f6fc",
		Text:       "#8b949e",
		Accent:     "#2f81f7",
	}

	lightTheme := Theme{
		Name:       "light",
		Background: "#ffffff",
		Border:     "#d0d7de",
		Title:      "#1f2328",
		Text:       "#656d76",
		Accent:     "#0969da",
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
