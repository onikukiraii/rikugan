package tui

import (
	"testing"

	"github.com/onikukiraii/rikugan/internal/diff"
)

func TestDiffSignature_same_input_same_hash(t *testing.T) {
	files := []diff.DiffFile{
		{
			OldName: "a.go", NewName: "a.go",
			Hunks: []diff.Hunk{
				{OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3,
					Lines: []diff.DiffLine{
						{Type: diff.LineContext, Content: "x", OldNum: 1, NewNum: 1},
					}},
			},
		},
	}

	sig1 := diffSignature(files)
	sig2 := diffSignature(files)
	if sig1 != sig2 {
		t.Error("same input should produce same signature")
	}
}

func TestDiffSignature_different_content_different_hash(t *testing.T) {
	files1 := []diff.DiffFile{
		{OldName: "a.go", NewName: "a.go",
			Hunks: []diff.Hunk{
				{OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
					Lines: []diff.DiffLine{
						{Type: diff.LineAdded, Content: "hello", NewNum: 1},
					}},
			}},
	}
	files2 := []diff.DiffFile{
		{OldName: "a.go", NewName: "a.go",
			Hunks: []diff.Hunk{
				{OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
					Lines: []diff.DiffLine{
						{Type: diff.LineAdded, Content: "world", NewNum: 1},
					}},
			}},
	}

	sig1 := diffSignature(files1)
	sig2 := diffSignature(files2)
	if sig1 == sig2 {
		t.Error("different content should produce different signatures")
	}
}

func TestDiffSignature_empty_files(t *testing.T) {
	sig := diffSignature(nil)
	if sig == "" {
		t.Error("signature of empty files should not be empty string")
	}
	// Should be consistent
	sig2 := diffSignature([]diff.DiffFile{})
	if sig != sig2 {
		t.Error("nil and empty slice should produce same signature")
	}
}

func TestDiffSignature_different_file_names(t *testing.T) {
	make := func(name string) []diff.DiffFile {
		return []diff.DiffFile{{OldName: name, NewName: name}}
	}
	sig1 := diffSignature(make("a.go"))
	sig2 := diffSignature(make("b.go"))
	if sig1 == sig2 {
		t.Error("different file names should produce different signatures")
	}
}

func TestDiffSignature_different_line_types(t *testing.T) {
	make := func(lt diff.LineType) []diff.DiffFile {
		return []diff.DiffFile{
			{OldName: "a.go", NewName: "a.go",
				Hunks: []diff.Hunk{
					{OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
						Lines: []diff.DiffLine{{Type: lt, Content: "x"}}},
				}},
		}
	}
	sigAdded := diffSignature(make(diff.LineAdded))
	sigRemoved := diffSignature(make(diff.LineRemoved))
	if sigAdded == sigRemoved {
		t.Error("different line types should produce different signatures")
	}
}
