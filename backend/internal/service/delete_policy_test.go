package service

import "testing"

func TestPlanDeckDeletion(t *testing.T) {
	cases := []struct {
		name             string
		visibility       string
		archived         bool
		removeFromMarket bool
		want             DeleteDeckMode
	}{
		{"public keep market", "public", false, false, DeleteArchived},
		{"public remove market", "public", false, true, DeleteHard},
		{"private ignores keep flag", "private", false, false, DeleteHard},
		{"private remove", "private", false, true, DeleteHard},
		{"archived always hard", "public", true, false, DeleteHard},
		{"archived remove", "public", true, true, DeleteHard},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := PlanDeckDeletion(c.visibility, c.archived, c.removeFromMarket); got != c.want {
				t.Fatalf("got %s, want %s", got, c.want)
			}
		})
	}
}
