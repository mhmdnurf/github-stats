package card

import (
	"math"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

type Rank struct {
	Level string
	Score float64
}

const (
	medianCommits      = 250.0
	medianPullRequests = 50.0
	medianStars        = 50.0
	medianFollowers    = 10.0
	medianRepositories = 25.0
)

const (
	weightCommits      = 2.0
	weightPullRequests = 1.5
	weightStars        = 2.0
	weightFollowers    = 1.0
	weightRepositories = 0.5
)

func computeRank(userStats stats.UserStats) Rank {
	score := weightCommits*expCDF(float64(userStats.Commits)/medianCommits) +
		weightPullRequests*expCDF(float64(userStats.PullRequests)/medianPullRequests) +
		weightStars*logisticCDF(float64(userStats.Stars)/medianStars) +
		weightFollowers*logisticCDF(float64(userStats.Followers)/medianFollowers) +
		weightRepositories*expCDF(float64(userStats.Repositories)/medianRepositories)

	totalWeight := weightCommits +
		weightPullRequests +
		weightStars +
		weightFollowers +
		weightRepositories

	normalized := score / totalWeight

	return Rank{
		Level: rankLevel(normalized),
		Score: normalized,
	}
}

func expCDF(x float64) float64 {
	if x < 0 {
		return 0
	}
	return 1 - math.Pow(2, -x)
}

func logisticCDF(x float64) float64 {
	if x < 0 {
		return 0
	}
	return x / (1 + x)
}

func rankLevel(score float64) string {
	switch {
	case score >= 0.90:
		return "S"
	case score >= 0.80:
		return "A+"
	case score >= 0.65:
		return "A"
	case score >= 0.50:
		return "B+"
	case score >= 0.35:
		return "B"
	case score >= 0.20:
		return "C+"
	default:
		return "C"
	}
}
