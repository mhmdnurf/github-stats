package card

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

const svgTemplate = `<svg
	xmlns="http://www.w3.org/2000/svg"
	width="495"
	height="195"
	viewBox="0 0 495 195"
	role="img"
	aria-label="{{.DisplayName}} GitHub statistics"
>
	<title>{{.DisplayName}} GitHub statistics</title>

	<rect
		x="0.5"
		y="0.5"
		width="494"
		height="194"
		rx="8"
		fill="{{.Theme.Background}}"
		stroke="{{.Theme.Border}}"
	/>

	<style>
		.header {
			font: 600 20px -apple-system, BlinkMacSystemFont,
				"Segoe UI", sans-serif;
		}

		.username, .label {
			font: 400 12px -apple-system, BlinkMacSystemFont,
				"Segoe UI", sans-serif;
		}

		.value {
			font: 600 18px -apple-system, BlinkMacSystemFont,
				"Segoe UI", sans-serif;
		}
	</style>

	<text
		x="24"
		y="34"
		class="header"
		fill="{{.Theme.Title}}"
	>{{.DisplayName}}</text>

	<text
		x="24"
		y="56"
		class="username"
		fill="{{.Theme.Text}}"
	>@{{.Stats.Username}}</text>

	<line
		x1="24"
		y1="70"
		x2="471"
		y2="70"
		stroke="{{.Theme.Border}}"
	/>

	<g>
		<text x="24" y="98" class="label" fill="{{.Theme.Text}}">Repositories</text>
		<text x="24" y="122" class="value" fill="{{.Theme.Accent}}">{{.Stats.Repositories}}</text>
		<text x="180" y="98" class="label" fill="{{.Theme.Text}}">Stars</text>
		<text x="180" y="122" class="value" fill="{{.Theme.Accent}}">{{.Stats.Stars}}</text>
		<text x="335" y="98" class="label" fill="{{.Theme.Text}}">Commits</text>
		<text x="335" y="122" class="value" fill="{{.Theme.Accent}}">{{.Stats.Commits}}</text>
		<text x="105" y="151" class="label" fill="{{.Theme.Text}}">Pull Requests</text>
		<text x="105" y="175" class="value" fill="{{.Theme.Accent}}">{{.Stats.PullRequests}}</text>
		<text x="310" y="151" class="label" fill="{{.Theme.Text}}">Followers</text>
		<text x="310" y="175" class="value" fill="{{.Theme.Accent}}">{{.Stats.Followers}}</text>
	</g>
</svg>`

type Renderer struct {
	template *template.Template
}

type renderData struct {
	Stats       stats.UserStats
	Theme       Theme
	DisplayName string
}

func NewRenderer() (*Renderer, error) {
	parsed, err := template.New("github-stats-card").Parse(svgTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse SVG template: %w", err)
	}

	return &Renderer{
		template: parsed,
	}, nil
}

func (renderer *Renderer) Render(
	userStats stats.UserStats,
	themeName string,
) ([]byte, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, fmt.Errorf("resolve card theme: %w", err)
	}

	data := renderData{
		Stats:       userStats,
		Theme:       theme,
		DisplayName: displayName(userStats),
	}

	var output bytes.Buffer
	if err := renderer.template.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("render SVG card: %w", err)
	}

	return output.Bytes(), nil
}

func displayName(userStats stats.UserStats) string {
	name := strings.TrimSpace(userStats.Name)
	if name == "" {
		name = userStats.Username
	}

	return truncateRunes(name, 32)
}

func truncateRunes(value string, maximum int) string {
	if maximum <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}

	return string(runes[:maximum-1]) + "…"
}
