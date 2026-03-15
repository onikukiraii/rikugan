package internal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/onikukiraii/rikugan/internal/diff"
)

// CopyToClipboard formats the diff with comments and copies to clipboard.
func CopyToClipboard(files []diff.DiffFile, comments map[diff.LineKey]string) error {
	text := FormatForAI(files, comments)
	return clipboard.WriteAll(text)
}

// CopyCommentsOnly copies only the comment summary to clipboard.
func CopyCommentsOnly(files []diff.DiffFile, comments map[diff.LineKey]string) error {
	text := FormatCommentsOnly(files, comments)
	return clipboard.WriteAll(text)
}

// FormatCommentsOnly generates a compact summary of comments without the full diff.
func FormatCommentsOnly(files []diff.DiffFile, comments map[diff.LineKey]string) string {
	if len(comments) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Review Comments\n\n")

	for key, comment := range comments {
		f := files[key.FileIndex]
		h := f.Hunks[key.HunkIndex]
		line := h.Lines[key.LineIndex]
		lineNum := line.NewNum
		if lineNum == 0 {
			lineNum = line.OldNum
		}

		if f.IsFullFile() {
			sb.WriteString(fmt.Sprintf("- **%s:%d** `%s`\n  %s\n", f.DisplayName(), lineNum, line.Content, comment))
		} else {
			var prefix string
			switch line.Type {
			case diff.LineAdded:
				prefix = "+"
			case diff.LineRemoved:
				prefix = "-"
			default:
				prefix = " "
			}
			sb.WriteString(fmt.Sprintf("- **%s:%d** `%s%s`\n  %s\n", f.DisplayName(), lineNum, prefix, line.Content, comment))
		}
	}

	return sb.String()
}

// FormatForAI generates a review-ready text with inline comments.
func FormatForAI(files []diff.DiffFile, comments map[diff.LineKey]string) string {
	var sb strings.Builder

	sb.WriteString("# Code Review\n\n")
	sb.WriteString("Please review the following diff with my inline comments.\n\n")

	for fi, f := range files {
		sb.WriteString(fmt.Sprintf("## %s\n\n", f.DisplayName()))

		if f.IsFullFile() {
			lang := extToLang(f.DisplayName())
			sb.WriteString(fmt.Sprintf("```%s\n", lang))
			for hi, h := range f.Hunks {
				for li, line := range h.Lines {
					sb.WriteString(line.Content + "\n")
					key := diff.LineKey{FileIndex: fi, HunkIndex: hi, LineIndex: li}
					if comment, ok := comments[key]; ok {
						sb.WriteString(fmt.Sprintf("// >> COMMENT: %s\n", comment))
					}
				}
			}
		} else {
			sb.WriteString("```diff\n")
			for hi, h := range f.Hunks {
				sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount))
				if h.Header != "" {
					sb.WriteString(" " + h.Header)
				}
				sb.WriteString("\n")

				for li, line := range h.Lines {
					var prefix string
					switch line.Type {
					case diff.LineAdded:
						prefix = "+"
					case diff.LineRemoved:
						prefix = "-"
					default:
						prefix = " "
					}
					sb.WriteString(fmt.Sprintf("%s%s\n", prefix, line.Content))

					key := diff.LineKey{FileIndex: fi, HunkIndex: hi, LineIndex: li}
					if comment, ok := comments[key]; ok {
						sb.WriteString(fmt.Sprintf("# >> COMMENT: %s\n", comment))
					}
				}
			}
		}
		sb.WriteString("```\n\n")
	}

	if len(comments) > 0 {
		sb.WriteString("## Summary of Comments\n\n")
		for key, comment := range comments {
			f := files[key.FileIndex]
			h := f.Hunks[key.HunkIndex]
			line := h.Lines[key.LineIndex]
			lineNum := line.NewNum
			if lineNum == 0 {
				lineNum = line.OldNum
			}
			sb.WriteString(fmt.Sprintf("- **%s:%d** - %s\n", f.DisplayName(), lineNum, comment))
		}
	}

	return sb.String()
}

// extToLang maps a file extension to a code fence language identifier.
func extToLang(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "go":
		return "go"
	case "py":
		return "python"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "jsx":
		return "jsx"
	case "tsx":
		return "tsx"
	case "rb":
		return "ruby"
	case "rs":
		return "rust"
	case "sh", "bash", "zsh":
		return "bash"
	case "yml", "yaml":
		return "yaml"
	case "md":
		return "markdown"
	case "":
		return ""
	default:
		return ext
	}
}
