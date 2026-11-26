package domain

type LeaderboardEntry struct {
	Rank         int    `json:"rank"`
	UserID       string `json:"userId"`
	Name         string `json:"name"`
	FormerMember bool   `json:"formerMember"`
	Count        int    `json:"count"`
}

type GetLeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}
