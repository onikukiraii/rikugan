package tui

import (
	"crypto/sha256"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/onikukiraii/rikugan/internal/diff"
)

// DiffLoader knows how to load diff files.
type DiffLoader struct {
	UseShow          bool
	Args             []string
	IncludeUntracked bool
}

// Load executes the diff command and returns the files.
func (l DiffLoader) Load() ([]diff.DiffFile, error) {
	var files []diff.DiffFile
	var err error
	if l.UseShow {
		files, err = diff.Show(l.Args)
	} else {
		files, err = diff.Run(l.Args)
	}
	if err != nil {
		return nil, err
	}
	if l.IncludeUntracked {
		untracked, utErr := diff.UntrackedFiles()
		if utErr == nil && len(untracked) > 0 {
			files = append(files, untracked...)
		}
	}
	return files, nil
}

// fileCheckMsg is sent by the background watcher when a periodic check completes.
type fileCheckMsg struct {
	files []diff.DiffFile
	sig   string
}

// reloadResultMsg is sent when a manual reload completes.
type reloadResultMsg struct {
	files []diff.DiffFile
	sig   string
}

func watchForChanges(loader DiffLoader, currentSig string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)
		files, err := loader.Load()
		if err != nil {
			return fileCheckMsg{sig: currentSig}
		}
		return fileCheckMsg{files: files, sig: diffSignature(files)}
	}
}

func manualReload(loader DiffLoader) tea.Cmd {
	return func() tea.Msg {
		files, err := loader.Load()
		if err != nil {
			return reloadResultMsg{}
		}
		return reloadResultMsg{files: files, sig: diffSignature(files)}
	}
}

func diffSignature(files []diff.DiffFile) string {
	h := sha256.New()
	for _, f := range files {
		h.Write([]byte(f.OldName))
		h.Write([]byte(f.NewName))
		for _, hk := range f.Hunks {
			fmt.Fprintf(h, "%d%d%d%d", hk.OldStart, hk.OldCount, hk.NewStart, hk.NewCount)
			for _, l := range hk.Lines {
				h.Write([]byte{byte(l.Type)})
				h.Write([]byte(l.Content))
			}
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
