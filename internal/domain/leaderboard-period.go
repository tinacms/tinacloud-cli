package domain

import "fmt"

type LeaderboardPeriod int

const (
	ThisMonth LeaderboardPeriod = iota
	LastMonth
	ThreeMonths
	TwelveMonths
)

func LeaderboardPeriodFromString(period string) (LeaderboardPeriod, error) {
	switch period {
	case "this-month":
		return ThisMonth, nil
	case "last-month":
		return LastMonth, nil
	case "3-months":
		return ThreeMonths, nil
	case "12-months":
		return TwelveMonths, nil
	}
	return 0, fmt.Errorf("invalid leaderboard period %q: use this-month, last-month, 3-months, or 12-months", period)
}
