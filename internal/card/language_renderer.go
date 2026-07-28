package card

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/languages"
)

type LanguageRenderer struct {
	template *template.Template
}

type languageRenderData struct {
	Username        string
	RepositoryLabel string
	Theme           Theme
	HasRows         bool
	Rows            []languageRow
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

func (renderer *LanguageRenderer) SupportsTheme(themeName string) bool {
	return SupportsTheme(themeName)
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
		RepositoryLabel: languageRepositoryLabel(userLanguages.Scope),
		Theme:           theme,
		HasRows:         len(rows) > 0,
		Rows:            rows,
	}

	var output bytes.Buffer
	if err := renderer.template.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("render language SVG card: %w", err)
	}

	return output.Bytes(), nil
}
