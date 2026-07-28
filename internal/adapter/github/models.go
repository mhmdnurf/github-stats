package github

type graphqlUser struct {
	Name                    *string              `json:"name"`
	Login                   string               `json:"login"`
	Repositories            repositoryConnection `json:"repositories"`
	ContributionsCollection contributionStats    `json:"contributionsCollection"`
	PullRequests            countConnection      `json:"pullRequests"`
	Followers               countConnection      `json:"followers"`
}

type repositoryConnection struct {
	TotalCount int          `json:"totalCount"`
	Nodes      []repository `json:"nodes"`
	PageInfo   pageInfo     `json:"pageInfo"`
}

type repository struct {
	StargazerCount int                `json:"stargazerCount"`
	IsFork         bool               `json:"isFork"`
	IsArchived     bool               `json:"isArchived"`
	Languages      languageConnection `json:"languages"`
}

type languageConnection struct {
	Edges []languageEdge `json:"edges"`
}

type languageEdge struct {
	Size int64        `json:"size"`
	Node languageNode `json:"node"`
}

type languageNode struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type pageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

type contributionStats struct {
	TotalCommitContributions int                  `json:"totalCommitContributions"`
	ContributionCalendar     contributionCalendar `json:"contributionCalendar"`
}

type contributionCalendar struct {
	Weeks []contributionWeek `json:"weeks"`
}

type contributionWeek struct {
	ContributionDays []contributionDay `json:"contributionDays"`
}

type contributionDay struct {
	ContributionCount int `json:"contributionCount"`
}

func (calendar contributionCalendar) weeklyTotals() []int {
	if len(calendar.Weeks) == 0 {
		return nil
	}

	totals := make([]int, 0, len(calendar.Weeks))
	for _, week := range calendar.Weeks {
		total := 0
		for _, day := range week.ContributionDays {
			total += day.ContributionCount
		}
		totals = append(totals, total)
	}

	return totals
}

type countConnection struct {
	TotalCount int `json:"totalCount"`
}
