package card

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

type Renderer struct {
	template *template.Template
}

type renderData struct {
	Stats       stats.UserStats
	Theme       Theme
	DisplayName string
	Rank        Rank
	Ring        ringData
	Spark       sparkData
	Cells       []cellData
	StatDelays  []string
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

func (renderer *Renderer) SupportsTheme(themeName string) bool {
	return SupportsTheme(themeName)
}

func (renderer *Renderer) Render(
	userStats stats.UserStats,
	themeName string,
) ([]byte, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, fmt.Errorf("resolve card theme: %w", err)
	}

	rank := computeRank(userStats)

	data := renderData{
		Stats:       userStats,
		Theme:       theme,
		DisplayName: displayName(userStats),
		Rank:        rank,
		Ring:        buildRing(rank.Score),
		Spark:       buildSparkline(userStats.WeeklyActivity),
		Cells:       buildCells(userStats),
		StatDelays:  statDelays(),
	}

	var output bytes.Buffer
	if err := renderer.template.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("render SVG card: %w", err)
	}

	return output.Bytes(), nil
}
