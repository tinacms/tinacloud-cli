package domain

type LeaderboardPeriod int

const (
	ThisMonth LeaderboardPeriod = iota
	LastMonth
	ThreeMonths
	TwelveMonths
)

func LeaderboardPeriodFromString(period string) LeaderboardPeriod {
	switch period {
	case "this-month":
		return ThisMonth
	case "last-month":
		return LastMonth
	case "3-months":
		return ThreeMonths
	case "12-months":
		return TwelveMonths
	}
	return ThisMonth
}
