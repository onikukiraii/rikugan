package diff

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFileSize  = 1 << 20 // 1MB
	maxFileCount = 1000
)

var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".venv":        true,
	".tox":         true,
	"dist":         true,
	"build":        true,
}

// FileToDisplayFile reads a file and converts it into a DiffFile where every
// line is a LineContext line in a single Hunk. This allows reuse of the
// existing inline/split view and comment infrastructure.
func FileToDisplayFile(path string) (DiffFile, error) {
	lines, err := ReadFileLines(path)
	if err != nil {
		return DiffFile{}, err
	}

	hunk := Hunk{
		OldStart: 0,
		OldCount: 0,
		NewStart: 1,
		NewCount: len(lines),
	}
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", "    ")
		hunk.Lines = append(hunk.Lines, DiffLine{
			Type:    LineContext,
			Content: line,
			NewNum:  i + 1,
		})
	}

	return DiffFile{
		NewName: path,
		Hunks:   []Hunk{hunk},
	}, nil
}

// ScanDirectory recursively walks dir and returns a DiffFile for each
// readable text file found, skipping hidden directories, known non-source
// directories, binary files, and files larger than 1 MB.
func ScanDirectory(dir string) ([]DiffFile, error) {
	var files []DiffFile
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}
		if len(files) >= maxFileCount {
			return filepath.SkipAll
		}
		if info.Size() > maxFileSize {
			return nil
		}
		if isBinaryExt(path) {
			return nil
		}
		df, err := FileToDisplayFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		files = append(files, df)
		return nil
	})
	return files, err
}

func isBinaryExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".webp",
		".mp3", ".mp4", ".wav", ".avi", ".mov",
		".zip", ".tar", ".gz", ".bz2", ".xz", ".7z",
		".exe", ".dll", ".so", ".dylib",
		".pdf", ".woff", ".woff2", ".ttf", ".otf",
		".bin", ".dat", ".o", ".a":
		return true
	}
	return false
}
