package domain

import "testing"

func TestLeaderboardPeriodFromString(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  LeaderboardPeriod
	}{
		{input: "this-month", want: ThisMonth},
		{input: "last-month", want: LastMonth},
		{input: "3-months", want: ThreeMonths},
		{input: "12-months", want: TwelveMonths},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := LeaderboardPeriodFromString(tc.input)
			if err != nil || got != tc.want {
				t.Errorf("LeaderboardPeriodFromString(%q) = %v, %v; want %v, nil", tc.input, got, err, tc.want)
			}
		})
	}
	for _, input := range []string{"", "this-mont"} {
		if _, err := LeaderboardPeriodFromString(input); err == nil {
			t.Errorf("LeaderboardPeriodFromString(%q) succeeded, want error", input)
		}
	}
}
