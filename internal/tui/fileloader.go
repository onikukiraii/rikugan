package tui

import "github.com/onikukiraii/rikugan/internal/diff"

// FileViewerLoader loads plain files (not git diffs) for viewing.
type FileViewerLoader struct {
	Dir   string
	Files []string
}

// Load reads the specified files or scans the directory and returns them
// as DiffFile structs suitable for the existing viewer.
func (l FileViewerLoader) Load() ([]diff.DiffFile, error) {
	if l.Dir != "" {
		return diff.ScanDirectory(l.Dir)
	}
	var files []diff.DiffFile
	for _, path := range l.Files {
		df, err := diff.FileToDisplayFile(path)
		if err != nil {
			return nil, err
		}
		files = append(files, df)
	}
	return files, nil
}
