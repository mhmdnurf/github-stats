package card

import (
	"fmt"
	"math"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	ringRadius        = 27
	ringStrokeWidth   = 5
	ringCircumference = 2 * math.Pi * ringRadius
)

type ringData struct {
	Circumference string
	Offset        string
}

type cellData struct {
	Label      string
	Value      string
	IconPath   string
	X          int
	LabelX     int
	IconY      int
	LabelY     int
	ValueY     int
	DelayIndex int
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

	cells := []cellData{
		{
			Label:      "STARS",
			Value:      formatCompact(userStats.Stars),
			IconPath:   iconStar,
			X:          columnOne,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 2,
		},
		{
			Label:      "COMMITS",
			Value:      formatCompact(userStats.Commits),
			IconPath:   iconCommit,
			X:          columnTwo,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 3,
		},
		{
			Label:      "PULL REQUESTS",
			Value:      formatCompact(userStats.PullRequests),
			IconPath:   iconPullRequest,
			X:          columnThree,
			LabelY:     rowOneLabelY,
			ValueY:     rowOneValueY,
			DelayIndex: 4,
		},
		{
			Label:      "REPOSITORIES",
			Value:      formatCompact(userStats.Repositories),
			IconPath:   iconRepo,
			X:          columnOne,
			LabelY:     rowTwoLabelY,
			ValueY:     rowTwoValueY,
			DelayIndex: 5,
		},
		{
			Label:      "FOLLOWERS",
			Value:      formatCompact(userStats.Followers),
			IconPath:   iconPeople,
			X:          columnTwo,
			LabelY:     rowTwoLabelY,
			ValueY:     rowTwoValueY,
			DelayIndex: 6,
		},
	}

	const (
		iconTextGap = 18
		iconAscent  = 10
	)

	for index := range cells {
		cells[index].LabelX = cells[index].X + iconTextGap
		cells[index].IconY = cells[index].LabelY - iconAscent
	}

	return cells
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
