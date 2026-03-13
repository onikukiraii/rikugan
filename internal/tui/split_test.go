package tui

import (
	"testing"

	"github.com/onikukiraii/rikugan/internal/diff"
)

func TestSplitBuildLines_fold_between_hunks(t *testing.T) {
	m := NewSplitModel()
	file := testFile() // reuse from inline_test.go
	m.BuildLines(file, 0, nil, -1)

	// Should have fold lines in both panes
	var foldFound bool
	for i := range m.leftPane {
		if m.leftPane[i].isFold && m.leftPane[i].foldIndex == 1 {
			foldFound = true
			if m.leftPane[i].foldLines != 12 {
				t.Errorf("left pane fold: expected 12 hidden lines, got %d", m.leftPane[i].foldLines)
			}
			// Right pane should match
			if !m.rightPane[i].isFold {
				t.Error("right pane should also have fold at same index")
			}
			break
		}
	}
	if !foldFound {
		t.Fatal("expected fold between hunks in split view")
	}
}

func TestSplitBuildLines_fold_commentable(t *testing.T) {
	m := NewSplitModel()
	file := testFile()
	m.BuildLines(file, 0, nil, -1)

	// Find the fold line
	for i := range m.leftPane {
		if m.leftPane[i].isFold {
			m.cursor = i
			break
		}
	}

	// Both panes should return valid key
	m.activePane = 0
	key, ok := m.CurrentLineKey()
	if !ok {
		t.Fatal("fold line in left pane should return valid key")
	}
	if key.HunkIndex != -1 {
		t.Errorf("expected HunkIndex -1, got %d", key.HunkIndex)
	}

	m.activePane = 1
	key2, ok2 := m.CurrentLineKey()
	if !ok2 {
		t.Fatal("fold line in right pane should return valid key")
	}
	if key != key2 {
		t.Error("both panes should return same key for fold line")
	}
}

func TestSplitBuildLines_expanded_fold(t *testing.T) {
	m := NewSplitModel()
	file := testFile()

	expanded := map[int][]diff.DiffLine{
		0: {
			{Type: diff.LineContext, Content: "l1", OldNum: 1, NewNum: 1},
			{Type: diff.LineContext, Content: "l2", OldNum: 2, NewNum: 2},
		},
	}
	m.BuildLines(file, 0, expanded, -1)

	// First two lines should be context (expanded), not fold
	if m.leftPane[0].isFold {
		t.Error("first line should be expanded context, not fold")
	}
	if m.leftPane[0].lineNum != 1 {
		t.Errorf("expected old lineNum 1, got %d", m.leftPane[0].lineNum)
	}
	if m.rightPane[0].lineNum != 1 {
		t.Errorf("expected new lineNum 1, got %d", m.rightPane[0].lineNum)
	}
	if m.leftPane[1].lineNum != 2 {
		t.Errorf("expected old lineNum 2, got %d", m.leftPane[1].lineNum)
	}
}

func TestSplitBuildLines_pane_lengths_match(t *testing.T) {
	m := NewSplitModel()
	file := testFile()
	m.BuildLines(file, 0, nil, 50)

	if len(m.leftPane) != len(m.rightPane) {
		t.Errorf("pane length mismatch: left=%d right=%d", len(m.leftPane), len(m.rightPane))
	}
}
