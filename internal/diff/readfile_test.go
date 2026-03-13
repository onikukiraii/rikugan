package diff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := ReadFileLines(path)
	if err != nil {
		t.Fatalf("ReadFileLines failed: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestReadFileLines_no_trailing_newline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("a\nb"), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := ReadFileLines(path)
	if err != nil {
		t.Fatalf("ReadFileLines failed: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestReadFileLines_empty_file(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := ReadFileLines(path)
	if err != nil {
		t.Fatalf("ReadFileLines failed: %v", err)
	}
	if lines != nil {
		t.Errorf("expected nil for empty file, got %v", lines)
	}
}

func TestReadFileLines_nonexistent(t *testing.T) {
	_, err := ReadFileLines("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
