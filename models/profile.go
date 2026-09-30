package models

// Activities are the "now" values in the home page now ledger.
// Leave a field empty to hide its line rather than showing a placeholder.
type Activities struct {
	Building   string
	Practicing string
	Playing    string
}

// CurrentActivities is what Daniel is building, practicing and playing now.
func CurrentActivities() Activities {
	return Activities{
		Building:   "this redesign",
		Practicing: "Allegro from Six Light Keyboard Pieces, Op. 52, No. 2",
		Playing:    "007 First Light on PS5",
	}
}
