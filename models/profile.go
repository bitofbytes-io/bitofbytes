package models

// Activities are the "now" values in the home page now ledger, written as
// displayed (first letter capitalized). Leave a field empty to hide its line
// rather than showing a placeholder.
type Activities struct {
	Building   string
	Practicing string
	Playing    string
}

// CurrentActivities is what Daniel is building, practicing and playing now.
func CurrentActivities() Activities {
	return Activities{
		Building:   "Trailmix, a new hiking app",
		Practicing: "Sonatina in E-flat Major, Op. 37, No. 1, first movement, by Muzio Clementi",
		Playing:    "Final Fantasy Resonance on Switch",
	}
}
