package internal

import (
	"strings"
	"testing"

	"github.com/onikukiraii/rikugan/internal/diff"
)

func TestFormatForAI(t *testing.T) {
	files := []diff.DiffFile{
		{
			OldName: "test.go",
			NewName: "test.go",
			Hunks: []diff.Hunk{
				{
					OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3,
					Lines: []diff.DiffLine{
						{Type: diff.LineContext, Content: "package main", OldNum: 1, NewNum: 1},
						{Type: diff.LineRemoved, Content: "var x = 1", OldNum: 2},
						{Type: diff.LineAdded, Content: "var x = 2", NewNum: 2},
						{Type: diff.LineContext, Content: "", OldNum: 3, NewNum: 3},
					},
				},
			},
		},
	}

	comments := map[diff.LineKey]string{
		{FileIndex: 0, HunkIndex: 0, LineIndex: 2}: "Why was this changed?",
	}

	result := FormatForAI(files, comments)

	if !strings.Contains(result, "# Code Review") {
		t.Error("missing header")
	}
	if !strings.Contains(result, "test.go") {
		t.Error("missing filename")
	}
	if !strings.Contains(result, "+var x = 2") {
		t.Error("missing added line")
	}
	if !strings.Contains(result, "COMMENT: Why was this changed?") {
		t.Error("missing inline comment")
	}
	if !strings.Contains(result, "Summary of Comments") {
		t.Error("missing summary section")
	}
}
