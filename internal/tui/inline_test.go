package tui

import (
	"testing"

	"github.com/onikukiraii/rikugan/internal/diff"
)

func testFile() diff.DiffFile {
	return diff.DiffFile{
		OldName: "test.go",
		NewName: "test.go",
		Hunks: []diff.Hunk{
			{
				OldStart: 5, OldCount: 3, NewStart: 5, NewCount: 4,
				Lines: []diff.DiffLine{
					{Type: diff.LineContext, Content: "ctx1", OldNum: 5, NewNum: 5},
					{Type: diff.LineRemoved, Content: "old", OldNum: 6},
					{Type: diff.LineAdded, Content: "new1", NewNum: 6},
					{Type: diff.LineAdded, Content: "new2", NewNum: 7},
					{Type: diff.LineContext, Content: "ctx2", OldNum: 7, NewNum: 8},
				},
			},
			{
				OldStart: 20, OldCount: 3, NewStart: 21, NewCount: 3,
				Lines: []diff.DiffLine{
					{Type: diff.LineContext, Content: "a", OldNum: 20, NewNum: 21},
					{Type: diff.LineRemoved, Content: "b", OldNum: 21},
					{Type: diff.LineAdded, Content: "c", NewNum: 22},
					{Type: diff.LineContext, Content: "d", OldNum: 22, NewNum: 23},
				},
			},
		},
	}
}

func TestBuildLines_fold_before_first_hunk(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1)

	// First hunk starts at NewStart=5, so lines 1-4 are hidden
	if len(m.lines) == 0 {
		t.Fatal("expected lines to be built")
	}
	first := m.lines[0]
	if !first.isFold {
		t.Fatal("expected first line to be a fold")
	}
	if first.foldLines != 4 {
		t.Errorf("expected 4 hidden lines, got %d", first.foldLines)
	}
	if first.foldIndex != 0 {
		t.Errorf("expected foldIndex 0, got %d", first.foldIndex)
	}
}

func TestBuildLines_fold_between_hunks(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1)

	// Between hunk 0 (ends at new line 5+4-1=8) and hunk 1 (starts at new line 21)
	// Hidden lines: 9-20 = 12 lines
	var foldBetween *renderedLine
	for i := range m.lines {
		if m.lines[i].isFold && m.lines[i].foldIndex == 1 {
			foldBetween = &m.lines[i]
			break
		}
	}
	if foldBetween == nil {
		t.Fatal("expected fold between hunks")
	}
	if foldBetween.foldLines != 12 {
		t.Errorf("expected 12 hidden lines between hunks, got %d", foldBetween.foldLines)
	}
}

func TestBuildLines_fold_after_last_hunk(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	// Last hunk ends at new line 21+3-1=23, totalLines=50, so 27 hidden
	m.BuildLines(file, 0, nil, 50)

	last := m.lines[len(m.lines)-1]
	if !last.isFold {
		t.Fatal("expected last line to be a fold (after last hunk)")
	}
	if last.foldLines != 27 {
		t.Errorf("expected 27 hidden lines after last hunk, got %d", last.foldLines)
	}
	if last.foldIndex != 2 {
		t.Errorf("expected foldIndex 2 (len(hunks)), got %d", last.foldIndex)
	}
}

func TestBuildLines_no_fold_after_last_hunk_when_unknown(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1) // totalLines unknown

	last := m.lines[len(m.lines)-1]
	if last.isFold {
		t.Error("should not have fold after last hunk when totalLines is unknown")
	}
}

func TestBuildLines_expanded_fold(t *testing.T) {
	m := NewInlineModel()
	file := testFile()

	// Expand fold 0 (before first hunk, lines 1-4)
	expanded := map[int][]diff.DiffLine{
		0: {
			{Type: diff.LineContext, Content: "line1", OldNum: 1, NewNum: 1},
			{Type: diff.LineContext, Content: "line2", OldNum: 2, NewNum: 2},
			{Type: diff.LineContext, Content: "line3", OldNum: 3, NewNum: 3},
			{Type: diff.LineContext, Content: "line4", OldNum: 4, NewNum: 4},
		},
	}
	m.BuildLines(file, 0, expanded, -1)

	// First 4 lines should be expanded context, not a fold
	for i := 0; i < 4; i++ {
		if m.lines[i].isFold {
			t.Errorf("line %d should be expanded context, not fold", i)
		}
		if m.lines[i].rawLine == nil {
			t.Errorf("line %d should have rawLine set", i)
		}
	}
	// Line 4 should be the hunk header
	if !m.lines[4].isHunk {
		t.Error("expected hunk header after expanded fold")
	}
}

func TestBuildLines_fold_has_commentable_key(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1)

	// Move cursor to the first fold line
	m.cursor = 0
	if !m.lines[0].isFold {
		t.Fatal("expected first line to be fold")
	}

	key, ok := m.CurrentLineKey()
	if !ok {
		t.Fatal("expected fold line to return a valid key for commenting")
	}
	if key.HunkIndex != -1 {
		t.Errorf("expected HunkIndex -1 for fold key, got %d", key.HunkIndex)
	}
}

func TestBuildLines_hunk_not_commentable(t *testing.T) {
	m := NewInlineModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1)

	// Find first hunk header
	for i, line := range m.lines {
		if line.isHunk {
			m.cursor = i
			break
		}
	}

	_, ok := m.CurrentLineKey()
	if ok {
		t.Error("hunk header should not be commentable")
	}
}

func TestBuildLines_no_fold_when_hunk_starts_at_line_1(t *testing.T) {
	m := NewInlineModel()
	file := diff.DiffFile{
		OldName: "a.go",
		NewName: "a.go",
		Hunks: []diff.Hunk{
			{
				OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3,
				Lines: []diff.DiffLine{
					{Type: diff.LineContext, Content: "pkg", OldNum: 1, NewNum: 1},
					{Type: diff.LineRemoved, Content: "x", OldNum: 2},
					{Type: diff.LineAdded, Content: "y", NewNum: 2},
					{Type: diff.LineContext, Content: "z", OldNum: 3, NewNum: 3},
				},
			},
		},
	}
	m.BuildLines(file, 0, nil, 3)

	// No fold before hunk (starts at line 1), no fold after (totalLines=3, hunk covers 1-3)
	for _, line := range m.lines {
		if line.isFold {
			t.Error("expected no fold lines when file fully covered by hunk")
		}
	}
}
