package card

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

func newTestLanguageRenderer(t *testing.T) *LanguageRenderer {
	t.Helper()

	renderer, err := NewLanguageRenderer()
	if err != nil {
		t.Fatalf("create language renderer: %v", err)
	}

	return renderer
}

func TestLanguageRendererRender(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	document, err := renderer.Render(
		languages.UserLanguages{
			Username: "mhmdnurf",
			Scope:    repositoryScope.ScopePublic,
			Languages: []languages.LanguageUsage{
				{
					Name:  "Go",
					Color: "#00ADD8",
					Bytes: 600,
				},
				{
					Name:  "TypeScript",
					Color: "#3178C6",
					Bytes: 250,
				},
				{
					Name:  "Python",
					Color: "#3572A5",
					Bytes: 100,
				},
				{
					Name:  "Rust",
					Color: "#DEA584",
					Bytes: 30,
				},
				{
					Name:  "Shell",
					Color: "invalid-color",
					Bytes: 15,
				},
				{
					Name:  "HTML",
					Color: "#E34C26",
					Bytes: 5,
				},
			},
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render language card: %v", err)
	}

	text := parseSVGText(t, document)

	for _, expected := range []string{
		"mhmdnurf most used programming languages",
		"Most Used Languages",
		"@mhmdnurf · public active repositories",
		"Go",
		"60.0%",
		"TypeScript",
		"25.0%",
		"Python",
		"10.0%",
		"Rust",
		"3.0%",
		"Shell",
		"1.5%",
	} {
		if !text[expected] {
			t.Errorf("expected SVG text %q", expected)
		}
	}

	if text["HTML"] {
		t.Fatal("expected the sixth language to be omitted")
	}

	output := string(document)

	for _, color := range []string{
		"#0d1117",
		"#30363d",
		"#f0f6fc",
		"#7d8590",
		"#e6edf3",
		"#21262d",
		"#00ADD8",
		"#3178C6",
		"#3572A5",
		"#DEA584",
		"#3fb950",
	} {
		if !strings.Contains(output, color) {
			t.Errorf("expected SVG color %q", color)
		}
	}

	if strings.Contains(output, "invalid-color") {
		t.Fatal("expected invalid language color to use the theme accent")
	}

	if !bytes.Contains(document, []byte(`width="147"`)) {
		t.Fatal("expected a 60 percent progress bar width")
	}
}

func TestLanguageRendererShowsAllRepositoryScope(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	document, err := renderer.Render(
		languages.UserLanguages{
			Username: "mhmdnurf",
			Scope:    repositoryScope.ScopeAll,
			Languages: []languages.LanguageUsage{
				{Name: "Go", Color: "#00ADD8", Bytes: 100},
			},
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render language card: %v", err)
	}

	text := parseSVGText(t, document)
	if !text["@mhmdnurf · all active repositories"] {
		t.Fatal("expected the all-repositories subtitle")
	}
}

func TestLanguageRendererShowsEmptyState(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	document, err := renderer.Render(
		languages.UserLanguages{
			Username: "mhmdnurf",
			Languages: []languages.LanguageUsage{
				{Name: " ", Bytes: 100},
				{Name: "Go", Bytes: 0},
				{Name: "Rust", Bytes: -1},
			},
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render language card: %v", err)
	}

	text := parseSVGText(t, document)

	if !text["No language data available"] {
		t.Fatal("expected the empty-state message")
	}

	if text["Go"] || text["Rust"] {
		t.Fatal("expected invalid language entries to be omitted")
	}
}

func TestLanguageRendererRejectsUnknownTheme(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	document, err := renderer.Render(
		languages.UserLanguages{Username: "mhmdnurf"},
		"unknown",
	)

	if !errors.Is(err, ErrUnknownTheme) {
		t.Fatalf("expected ErrUnknownTheme, got %v", err)
	}

	if document != nil {
		t.Fatalf("expected nil document, got %q", document)
	}
}

func TestLanguageRendererEscapesUntrustedText(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	document, err := renderer.Render(
		languages.UserLanguages{
			Username: `user<&`,
			Languages: []languages.LanguageUsage{
				{
					Name:  `<script>alert("unsafe")</script>`,
					Color: "#00ADD8",
					Bytes: 100,
				},
			},
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render language card: %v", err)
	}

	output := string(document)

	if strings.Contains(output, "<script>") {
		t.Fatal("rendered output contains an injected script element")
	}

	if !strings.Contains(output, "&lt;script&gt;") {
		t.Fatal("expected language name to be escaped")
	}

	if !strings.Contains(output, "user&lt;&amp;") {
		t.Fatal("expected username to be escaped")
	}

	parseSVGText(t, document)
}

func TestLanguageRendererProducesDeterministicOutput(t *testing.T) {
	renderer := newTestLanguageRenderer(t)

	input := languages.UserLanguages{
		Username: "mhmdnurf",
		Languages: []languages.LanguageUsage{
			{Name: "Go", Color: "#00ADD8", Bytes: 100},
			{Name: "Rust", Color: "#DEA584", Bytes: 100},
			{Name: "Python", Color: "#3572A5", Bytes: 50},
		},
	}

	first, err := renderer.Render(input, DefaultTheme)
	if err != nil {
		t.Fatalf("render first card: %v", err)
	}

	second, err := renderer.Render(input, DefaultTheme)
	if err != nil {
		t.Fatalf("render second card: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("expected deterministic SVG output")
	}
}
