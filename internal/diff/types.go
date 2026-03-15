package diff

// DiffFile represents a single file's diff.
type DiffFile struct {
	OldName string
	NewName string
	Hunks   []Hunk
}

// DisplayName returns the best name to display for this file.
func (f DiffFile) DisplayName() string {
	if f.NewName != "" && f.NewName != "/dev/null" {
		return f.NewName
	}
	return f.OldName
}

// Hunk represents a single diff hunk.
type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []DiffLine
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type    LineType
	Content string
	OldNum  int // 0 if not applicable
	NewNum  int // 0 if not applicable
}

// LineType indicates whether a line was added, removed, or unchanged.
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
)

// IsFullFile returns true if this DiffFile represents a plain file view
// (not a git diff), indicated by a single hunk starting at line 1 with no old name.
func (f DiffFile) IsFullFile() bool {
	return len(f.Hunks) == 1 && f.Hunks[0].NewStart == 1 && f.OldName == ""
}

// LineKey uniquely identifies a line in a diff for comment attachment.
type LineKey struct {
	FileIndex int
	HunkIndex int
	LineIndex int
}
