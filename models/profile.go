package models

// Activities are the "now" lines at the bottom of the home page hobby cards.
// Leave a field empty to hide its line rather than showing a placeholder.
type Activities struct {
	Building   string
	Practicing string
	Playing    string
}

// CurrentActivities is what Daniel is building, practicing and playing now.
func CurrentActivities() Activities {
	return Activities{
		Building: "this redesign",
	}
}
