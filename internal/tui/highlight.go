package tui

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter provides syntax highlighting for diff lines.
type Highlighter struct {
	lexer chroma.Lexer
	style *chroma.Style
}

// NewHighlighter creates a highlighter based on the filename.
func NewHighlighter(filename string) Highlighter {
	ext := filepath.Ext(filename)
	base := filepath.Base(filename)

	var lexer chroma.Lexer
	lexer = lexers.Match(base)
	if lexer == nil && ext != "" {
		lexer = lexers.Match("file" + ext)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	return Highlighter{lexer: lexer, style: style}
}

// Highlight applies syntax coloring to a line of code, returning an ANSI string.
func (h Highlighter) Highlight(code string) string {
	if code == "" {
		return ""
	}

	iter, err := h.lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var sb strings.Builder
	for _, tok := range iter.Tokens() {
		entry := h.style.Get(tok.Type)
		sb.WriteString(applyStyle(tok.Value, entry))
	}
	return sb.String()
}

func applyStyle(text string, entry chroma.StyleEntry) string {
	if !entry.Colour.IsSet() && entry.Bold != chroma.Yes && entry.Italic != chroma.Yes {
		return text
	}

	var codes []string

	if entry.Bold == chroma.Yes {
		codes = append(codes, "1")
	}
	if entry.Italic == chroma.Yes {
		codes = append(codes, "3")
	}
	if entry.Colour.IsSet() {
		r, g, b := entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue()
		codes = append(codes, rgbFg(r, g, b))
	}

	if len(codes) == 0 {
		return text
	}

	return "\033[" + strings.Join(codes, ";") + "m" + text + "\033[0m"
}

func rgbFg(r, g, b uint8) string {
	return "38;2;" + uitoa(r) + ";" + uitoa(g) + ";" + uitoa(b)
}

func uitoa(v uint8) string {
	if v < 10 {
		return string(rune('0' + v))
	}
	if v < 100 {
		return string([]byte{byte('0' + v/10), byte('0' + v%10)})
	}
	return string([]byte{byte('0' + v/100), byte('0' + (v/10)%10), byte('0' + v%10)})
}
