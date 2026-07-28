package card

import (
	"fmt"
	"math"
	"strings"
)

const (
	sparkLeft   = 352.0
	sparkRight  = 467.0
	sparkTop    = 172.0
	sparkBottom = 196.0
)

type sparkData struct {
	Show   bool
	Points string
	Length string
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
