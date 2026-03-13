package diff

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	godiff "github.com/sourcegraph/go-diff/diff"
)

// Run executes git diff with the given arguments and parses the output.
func Run(args []string) ([]DiffFile, error) {
	cmdArgs := append([]string{"diff"}, args...)
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return Parse(string(out))
}

// Show executes git show with the given commit and parses the diff output.
func Show(args []string) ([]DiffFile, error) {
	cmdArgs := append([]string{"show", "--format="}, args...)
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git show failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git show failed: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return Parse(string(out))
}

// IsCommit checks if the given ref resolves to a commit object.
func IsCommit(ref string) bool {
	cmd := exec.Command("git", "cat-file", "-t", ref)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "commit"
}

// UntrackedFiles returns diff representations of untracked files.
func UntrackedFiles() ([]DiffFile, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}

	var files []DiffFile
	for _, path := range strings.Split(text, "\n") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		// --no-index exits with code 1 when differences exist, but stdout
		// still contains the diff output. CombinedOutput captures it reliably.
		diffOut, _ := exec.Command("git", "diff", "--no-index", "--", "/dev/null", path).CombinedOutput()
		if len(diffOut) == 0 {
			continue
		}
		parsed, parseErr := Parse(string(diffOut))
		if parseErr != nil || len(parsed) == 0 {
			continue
		}
		for i := range parsed {
			if parsed[i].OldName == "dev/null" || parsed[i].OldName == "/dev/null" {
				parsed[i].OldName = "/dev/null"
			}
			parsed[i].NewName = path
		}
		files = append(files, parsed...)
	}
	return files, nil
}

// ReadFileLines reads a file and returns its lines.
func ReadFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(data)
	if text == "" {
		return nil, nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// Parse parses unified diff text into DiffFile structs.
func Parse(text string) ([]DiffFile, error) {
	fileDiffs, err := godiff.ParseMultiFileDiff([]byte(text))
	if err != nil {
		return nil, fmt.Errorf("parse diff: %w", err)
	}

	var files []DiffFile
	for _, fd := range fileDiffs {
		f := DiffFile{
			OldName: stripPrefix(fd.OrigName),
			NewName: stripPrefix(fd.NewName),
		}
		for _, h := range fd.Hunks {
			hunk := parseHunk(h)
			f.Hunks = append(f.Hunks, hunk)
		}
		files = append(files, f)
	}
	return files, nil
}

func parseHunk(h *godiff.Hunk) Hunk {
	hunk := Hunk{
		OldStart: int(h.OrigStartLine),
		OldCount: int(h.OrigLines),
		NewStart: int(h.NewStartLine),
		NewCount: int(h.NewLines),
		Header:   strings.TrimSpace(string(h.Section)),
	}

	body := string(h.Body)
	rawLines := strings.Split(body, "\n")
	// Remove trailing empty line from split
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	oldNum := int(h.OrigStartLine)
	newNum := int(h.NewStartLine)

	for _, line := range rawLines {
		if len(line) == 0 {
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    LineContext,
				Content: "",
				OldNum:  oldNum,
				NewNum:  newNum,
			})
			oldNum++
			newNum++
			continue
		}

		prefix := line[0]
		content := expandTabs(line[1:], 4)

		switch prefix {
		case '+':
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    LineAdded,
				Content: content,
				NewNum:  newNum,
			})
			newNum++
		case '-':
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    LineRemoved,
				Content: content,
				OldNum:  oldNum,
			})
			oldNum++
		default:
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    LineContext,
				Content: content,
				OldNum:  oldNum,
				NewNum:  newNum,
			})
			oldNum++
			newNum++
		}
	}
	return hunk
}

func expandTabs(s string, tabWidth int) string {
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
}

func stripPrefix(name string) string {
	name = strings.TrimPrefix(name, "a/")
	name = strings.TrimPrefix(name, "b/")
	return name
}
