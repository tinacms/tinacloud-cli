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
	Landing *Landing           `json:"landing,omitempty"`
}

type Landing struct {
	At           int64                `json:"at"`
	Branch       string               `json:"branch,omitempty"`
	Contributors []LandingContributor `json:"contributors"`
}

type LandingContributor struct {
	UserID       string  `json:"userId"`
	Name         *string `json:"name"`
	FormerMember bool    `json:"formerMember"`
}
