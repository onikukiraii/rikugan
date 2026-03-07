package tui

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	Quit       string
	Up         string
	Down       string
	HalfPageUp string
	HalfPageDn string
	Top        string
	Bottom     string
	NextFile   string
	PrevFile   string
	Comment    string
	DelComment string
	Copy       string
	ToggleMode string
	PaneLeft   string
	PaneRight  string
	Help       string
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:       "q",
		Up:         "k",
		Down:       "j",
		HalfPageUp: "ctrl+u",
		HalfPageDn: "ctrl+d",
		Top:        "g",
		Bottom:     "G",
		NextFile:   "tab",
		PrevFile:   "shift+tab",
		Comment:    "c",
		DelComment: "d",
		Copy:       "y",
		ToggleMode: "V",
		PaneLeft:   "h",
		PaneRight:  "l",
		Help:       "?",
	}
}
