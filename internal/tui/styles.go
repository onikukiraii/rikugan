package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	// Colors
	colorRed     = color.RGBA{0xFF, 0x6B, 0x6B, 0xFF}
	colorGreen   = color.RGBA{0x69, 0xDB, 0x7C, 0xFF}
	colorYellow  = color.RGBA{0xFF, 0xD4, 0x3B, 0xFF}
	colorBlue    = color.RGBA{0x74, 0xC0, 0xFC, 0xFF}
	colorDim     = color.RGBA{0x86, 0x86, 0x86, 0xFF}
	colorBg      = color.RGBA{0x1E, 0x1E, 0x2E, 0xFF}
	colorBgLight = color.RGBA{0x31, 0x31, 0x44, 0xFF}
	colorWhite   = color.RGBA{0xCD, 0xD6, 0xF4, 0xFF}
	colorOrange  = color.RGBA{0xFA, 0xB3, 0x87, 0xFF}
	colorMagenta = color.RGBA{0xCB, 0xA6, 0xF7, 0xFF}

	// Subtle background tints for diff lines
	colorBgAdded   = color.RGBA{0x1A, 0x2E, 0x1A, 0xFF} // dark green tint
	colorBgRemoved = color.RGBA{0x2E, 0x1A, 0x1A, 0xFF} // dark red tint

	// Line styles
	styleAdded = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleRemoved = lipgloss.NewStyle().
			Foreground(colorRed)

	styleContext = lipgloss.NewStyle().
			Foreground(colorDim)

	styleLineNum = lipgloss.NewStyle().
			Foreground(colorDim).
			Width(5).
			Align(lipgloss.Right)

	styleLineNumActive = lipgloss.NewStyle().
				Foreground(colorBlue).
				Width(5).
				Align(lipgloss.Right)

	styleLineNumAdded = lipgloss.NewStyle().
				Foreground(colorGreen).
				Width(5).
				Align(lipgloss.Right)

	styleLineNumRemoved = lipgloss.NewStyle().
				Foreground(colorRed).
				Width(5).
				Align(lipgloss.Right)

	// UI chrome
	styleHeader = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true)

	styleHunkHeader = lipgloss.NewStyle().
			Foreground(colorMagenta)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorBlue).
			Padding(0, 1)

	styleStatusBarSection = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorBgLight).
				Padding(0, 1)

	styleComment = lipgloss.NewStyle().
			Foreground(colorYellow).
			Italic(true)

	styleCommentIndicator = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true)

	styleFileTab = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	styleFileTabActive = lipgloss.NewStyle().
				Foreground(colorBlue).
				Bold(true).
				Padding(0, 1).
				Underline(true)

	styleCursorLine = lipgloss.NewStyle().
			Background(colorBgLight)

	// Line background styles (applied to entire line for diff visibility)
	styleBgAdded = lipgloss.NewStyle().
			Background(colorBgAdded)

	styleBgRemoved = lipgloss.NewStyle().
			Background(colorBgRemoved)

	// Word diff: highlight the specific characters that changed
	styleWordDiffAdded = lipgloss.NewStyle().
				Foreground(colorGreen).
				Background(color.RGBA{0x2A, 0x4A, 0x2A, 0xFF}).
				Bold(true)

	styleWordDiffRemoved = lipgloss.NewStyle().
				Foreground(colorRed).
				Background(color.RGBA{0x4A, 0x2A, 0x2A, 0xFF}).
				Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorDim)
)
