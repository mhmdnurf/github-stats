package card

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/languages"
)

const (
	languageLimit    = 5
	languageBarWidth = 245
	languageFirstY   = 91
	languageRowGap   = 28
)

type LanguageRenderer struct {
	template *template.Template
}

type languageRenderData struct {
	Username string
	Theme    Theme
	HasRows  bool
	Rows     []languageRow
}

type languageRow struct {
	Name       string
	Color      string
	Percentage string
	Width      int
	Y          int
	BarY       int
}

func NewLanguageRenderer() (*LanguageRenderer, error) {
	parsed, err := template.New("github-languages-card").Parse(
		languagesSVGTemplate,
	)
	if err != nil {
		return nil, fmt.Errorf("parse language SVG template: %w", err)
	}

	return &LanguageRenderer{
		template: parsed,
	}, nil
}

func (renderer *LanguageRenderer) Render(
	userLanguages languages.UserLanguages,
	themeName string,
) ([]byte, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, fmt.Errorf("resolve card theme: %w", err)
	}

	rows := buildLanguageRows(userLanguages.Languages, theme.Accent)

	data := languageRenderData{
		Username: truncateRunes(
			strings.TrimSpace(userLanguages.Username),
			39,
		),
		Theme:   theme,
		HasRows: len(rows) > 0,
		Rows:    rows,
	}

	var output bytes.Buffer
	if err := renderer.template.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("render language SVG card: %w", err)
	}

	return output.Bytes(), nil
}

func buildLanguageRows(
	usages []languages.LanguageUsage,
	fallbackColor string,
) []languageRow {
	totals := make(map[string]languages.LanguageUsage)

	for _, usage := range usages {
		name := strings.TrimSpace(usage.Name)
		if name == "" || usage.Bytes <= 0 {
			continue
		}

		total := totals[name]
		total.Name = name
		total.Bytes += usage.Bytes

		if total.Color == "" && validLanguageColor(usage.Color) {
			total.Color = usage.Color
		}

		totals[name] = total
	}

	totalBytes := int64(0)
	values := make(
		[]languages.LanguageUsage,
		0,
		len(totals),
	)

	for _, usage := range totals {
		totalBytes += usage.Bytes
		values = append(values, usage)
	}

	if totalBytes == 0 {
		return nil
	}

	sort.Slice(values, func(left, right int) bool {
		if values[left].Bytes == values[right].Bytes {
			return values[left].Name < values[right].Name
		}

		return values[left].Bytes > values[right].Bytes
	})

	if len(values) > languageLimit {
		values = values[:languageLimit]
	}

	rows := make([]languageRow, 0, len(values))

	for index, usage := range values {
		ratio := float64(usage.Bytes) / float64(totalBytes)
		width := int(math.Round(ratio * languageBarWidth))

		if width < 0 {
			width = 0
		}
		if width > languageBarWidth {
			width = languageBarWidth
		}

		color := fallbackColor
		if validLanguageColor(usage.Color) {
			color = usage.Color
		}

		y := languageFirstY + index*languageRowGap

		rows = append(rows, languageRow{
			Name:       truncateRunes(usage.Name, 20),
			Color:      color,
			Percentage: fmt.Sprintf("%.1f%%", ratio*100),
			Width:      width,
			Y:          y,
			BarY:       y - 4,
		})
	}

	return rows
}

func validLanguageColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}

	for _, character := range value[1:] {
		isDigit := character >= '0' && character <= '9'
		isLowerHex := character >= 'a' && character <= 'f'
		isUpperHex := character >= 'A' && character <= 'F'

		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}

	return true
}
