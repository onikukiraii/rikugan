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
	NextHunk   string
	PrevHunk   string
	Comment    string
	DelComment string
	Copy        string
	CopySummary string
	ToggleMode  string
	PaneLeft    string
	PaneRight   string
	ExpandFold  string
	Help        string
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
		NextHunk:   "]",
		PrevHunk:   "[",
		Comment:    "c",
		DelComment: "d",
		Copy:        "Y",
		CopySummary: "y",
		ToggleMode:  "V",
		PaneLeft:    "h",
		PaneRight:   "l",
		ExpandFold:  "enter",
		Help:        "?",
	}
}
