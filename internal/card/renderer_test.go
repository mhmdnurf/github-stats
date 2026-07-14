package card

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	return renderer
}

func parseSVGText(t *testing.T, document []byte) map[string]bool {
	t.Helper()

	values := make(map[string]bool)
	decoder := xml.NewDecoder(bytes.NewReader(document))

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("parse SVG: %v", err)
		}

		characters, ok := token.(xml.CharData)
		if !ok {
			continue
		}

		value := strings.TrimSpace(string(characters))
		if value != "" {
			values[value] = true
		}
	}

	return values
}

func TestRendererRender(t *testing.T) {
	renderer := newTestRenderer(t)

	userStats := stats.UserStats{
		Name:         "Muhammad Nurfatkhur Rahman",
		Username:     "mhmdnurf",
		Repositories: 101,
		Stars:        202,
		Commits:      303,
		PullRequests: 404,
		Followers:    505,
	}

	document, err := renderer.Render(userStats, DefaultTheme)
	if err != nil {
		t.Fatalf("render card: %v", err)
	}

	text := parseSVGText(t, document)

	expectedText := []string{
		"Muhammad Nurfatkhur Rahman GitHub statistics",
		"Muhammad Nurfatkhur Rahman",
		"@mhmdnurf",
		"Repositories",
		"101",
		"Stars",
		"202",
		"Commits",
		"303",
		"Pull Requests",
		"404",
		"Followers",
		"505",
	}

	for _, expected := range expectedText {
		if !text[expected] {
			t.Errorf("expected SVG text %q", expected)
		}
	}

	output := string(document)
	for _, color := range []string{
		"#0d1117",
		"#30363d",
		"#f0f6fc",
		"#8b949e",
		"#2f81f7",
	} {
		if !strings.Contains(output, color) {
			t.Errorf("expected default-theme color %q", color)
		}
	}
}

func TestRendererEscapesUntrustedText(t *testing.T) {
	renderer := newTestRenderer(t)

	document, err := renderer.Render(
		stats.UserStats{
			Name:     `<script>alert("unsafe")</script>`,
			Username: `user<&`,
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render card: %v", err)
	}

	output := string(document)

	if strings.Contains(output, "<script>") {
		t.Fatal("rendered output contains an injected script element")
	}

	if !strings.Contains(output, "&lt;script&gt;") {
		t.Fatal("expected profile name to be escaped")
	}

	if !strings.Contains(output, "user&lt;&amp;") {
		t.Fatal("expected username to be escaped")
	}

	// Also verifies that escaping still produces valid XML.
	parseSVGText(t, document)
}

func TestRendererUsesFallbackName(t *testing.T) {
	renderer := newTestRenderer(t)

	document, err := renderer.Render(
		stats.UserStats{
			Name:     "   ",
			Username: "mhmdnurf",
		},
		DefaultTheme,
	)
	if err != nil {
		t.Fatalf("render card: %v", err)
	}

	text := parseSVGText(t, document)
	if !text["mhmdnurf"] {
		t.Fatal("expected username as the display-name fallback")
	}
}

func TestRendererUsesSelectedTheme(t *testing.T) {
	renderer := newTestRenderer(t)

	document, err := renderer.Render(
		stats.UserStats{Username: "mhmdnurf"},
		LightTheme,
	)
	if err != nil {
		t.Fatalf("render card: %v", err)
	}

	output := string(document)

	for _, color := range []string{
		"#ffffff",
		"#d0d7de",
		"#1f2328",
		"#656d76",
		"#0969da",
	} {
		if !strings.Contains(output, color) {
			t.Errorf("expected light-theme color %q", color)
		}
	}
}

func TestRendererRejectsUnknownTheme(t *testing.T) {
	renderer := newTestRenderer(t)

	document, err := renderer.Render(
		stats.UserStats{Username: "mhmdnurf"},
		"unknown",
	)

	if !errors.Is(err, ErrUnknownTheme) {
		t.Fatalf("expected ErrUnknownTheme, got %v", err)
	}

	if document != nil {
		t.Fatalf("expected nil document, got %q", document)
	}
}

func TestRendererProducesDeterministicOutput(t *testing.T) {
	renderer := newTestRenderer(t)
	userStats := stats.UserStats{
		Name:     "Muhammad Nurfatkhur Rahman",
		Username: "mhmdnurf",
		Stars:    128,
	}

	first, err := renderer.Render(userStats, DefaultTheme)
	if err != nil {
		t.Fatalf("render first card: %v", err)
	}

	second, err := renderer.Render(userStats, DefaultTheme)
	if err != nil {
		t.Fatalf("render second card: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("expected deterministic SVG output")
	}
}

func TestTruncateRunes(t *testing.T) {
	input := strings.Repeat("界", 33)

	got := truncateRunes(input, 32)
	want := strings.Repeat("界", 31) + "…"

	if got != want {
		t.Fatalf("unexpected truncation: got %q, want %q", got, want)
	}

	if utf8.RuneCountInString(got) != 32 {
		t.Fatalf(
			"expected 32 runes, got %d",
			utf8.RuneCountInString(got),
		)
	}

	if got := truncateRunes("value", 0); got != "" {
		t.Fatalf("expected empty result for zero maximum, got %q", got)
	}
}
