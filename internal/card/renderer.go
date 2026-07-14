package card

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	cardWidth  = 495
	cardHeight = 240

	ringRadius        = 27
	ringStrokeWidth   = 5
	ringCircumference = 2 * math.Pi * ringRadius

	sparkLeft   = 352.0
	sparkRight  = 467.0
	sparkTop    = 172.0
	sparkBottom = 196.0
)

type Renderer struct {
	template *template.Template
}

type ringData struct {
	Circumference string
	Offset        string
}

type sparkData struct {
	Show   bool
	Points string
	Length string
}

type cellData struct {
	Label      string
	Value      string
	X          int
	LabelY     int
	ValueY     int
	DelayIndex int
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

func buildCells(userStats stats.UserStats) []cellData {
	const (
		columnOne   = 28
		columnTwo   = 190
		columnThree = 352

		rowOneLabelY = 104
		rowOneValueY = 130
		rowTwoLabelY = 164
		rowTwoValueY = 190
	)

	return []cellData{
		{
			Label:      "STARS",
			Value:      formatCompact(userStats.Stars),
			X:          columnOne,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 2,
		},
		{
			Label:      "COMMITS",
			Value:      formatCompact(userStats.Commits),
			X:          columnTwo,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 3,
		},
		{
			Label:      "PULL REQUESTS",
			Value:      formatCompact(userStats.PullRequests),
			X:          columnThree,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 4,
		},
		{
			Label:      "REPOSITORIES",
			Value:      formatCompact(userStats.Repositories),
			X:          columnOne,
			LabelY:     rowTwoLabelY,
			ValueY:     rowTwoValueY,
			DelayIndex: 5,
		},
		{
			Label:      "FOLLOWERS",
			Value:      formatCompact(userStats.Followers),
			X:          columnTwo,
			LabelY:     rowTwoLabelY,
			ValueY:     rowTwoValueY,
			DelayIndex: 6,
		},
	}
}

func statDelays() []string {
	delays := make([]string, 7)
	for index := range delays {
		delays[index] = fmt.Sprintf("%.2f", float64(index)*0.08)
	}
	return delays
}

func buildRing(score float64) ringData {
	clamped := math.Max(0, math.Min(1, score))
	offset := ringCircumference * (1 - clamped)

	return ringData{
		Circumference: fmt.Sprintf("%.2f", ringCircumference),
		Offset:        fmt.Sprintf("%.2f", offset),
	}
}

func buildSparkline(weeks []int) sparkData {
	if len(weeks) < 2 {
		return sparkData{Show: false}
	}

	maximum := 0
	for _, count := range weeks {
		if count > maximum {
			maximum = count
		}
	}

	width := sparkRight - sparkLeft
	height := sparkBottom - sparkTop
	step := width / float64(len(weeks)-1)

	points := make([]string, 0, len(weeks))
	coordinates := make([][2]float64, 0, len(weeks))

	for index, count := range weeks {
		x := sparkLeft + float64(index)*step

		y := sparkBottom
		if maximum > 0 {
			y = sparkBottom - (float64(count)/float64(maximum))*height
		}

		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
		coordinates = append(coordinates, [2]float64{x, y})
	}

	length := 0.0
	for index := 1; index < len(coordinates); index++ {
		deltaX := coordinates[index][0] - coordinates[index-1][0]
		deltaY := coordinates[index][1] - coordinates[index-1][1]
		length += math.Hypot(deltaX, deltaY)
	}

	return sparkData{
		Show:   true,
		Points: strings.Join(points, " "),
		Length: fmt.Sprintf("%.1f", math.Ceil(length)),
	}
}

func displayName(userStats stats.UserStats) string {
	name := strings.TrimSpace(userStats.Name)
	if name == "" {
		name = userStats.Username
	}

	return truncateRunes(name, 28)
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
