package card

import (
	"fmt"
	"strings"
)

// formatCompact renders a count the way GitHub does: 999 stays 999,
// 1204 becomes "1.2k", 12040 becomes "12k", 3400000 becomes "3.4M".
func formatCompact(value int) string {
	switch {
	case value < 1_000:
		return fmt.Sprintf("%d", value)
	case value < 100_000:
		return trimZero(fmt.Sprintf("%.1f", float64(value)/1_000)) + "k"
	case value < 1_000_000:
		return fmt.Sprintf("%dk", value/1_000)
	default:
		return trimZero(fmt.Sprintf("%.1f", float64(value)/1_000_000)) + "M"
	}
}

func trimZero(value string) string {
	return strings.TrimSuffix(value, ".0")
}
