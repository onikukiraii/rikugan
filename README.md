<p align="center">
  <img src="https://raw.githubusercontent.com/onikukiraii/rikugan/main/assets/logo.svg" alt="rikugan logo" width="120" height="120" />
</p>

<h1 align="center">rikugan 六眼</h1>

<p align="center">
  <strong>TUI diff reviewer for AI-powered code review</strong>
</p>

<p align="center">
  <a href="https://github.com/onikukiraii/rikugan/releases"><img src="https://img.shields.io/github/v/release/onikukiraii/rikugan?style=flat-square" alt="Release" /></a>
  <a href="https://github.com/onikukiraii/rikugan/blob/main/LICENSE"><img src="https://img.shields.io/github/license/onikukiraii/rikugan?style=flat-square" alt="License" /></a>
  <a href="https://github.com/onikukiraii/rikugan/actions"><img src="https://img.shields.io/github/actions/workflow/status/onikukiraii/rikugan/release.yml?style=flat-square" alt="CI" /></a>
</p>

<p align="center">
  Review git diffs interactively in your terminal.<br />
  Add inline comments, then copy the annotated diff to clipboard — ready for AI code review.
</p>

---

## Features

- **Inline & Split view** — Toggle between unified and side-by-side diff views with `V`
- **Syntax highlighting** — Language-aware coloring powered by [Chroma](https://github.com/alecthomas/chroma)
- **Word-level diff** — Character-granularity highlighting shows exactly what changed within each line
- **Inline comments** — Add review comments on any diff line, carried through to the clipboard output
- **Fuzzy file picker** — Press `Space` to fuzzy-search across changed files
- **AI-ready clipboard output** — Copy the full annotated diff or a comments-only summary, formatted for AI assistants
- **Hunk navigation** — Jump between diff hunks with `]` / `[`
- **Vim-style keybindings** — `j`/`k`, `gg`/`G`, `Ctrl+d`/`Ctrl+u` — feels like home

## Demo

<!-- TODO: Add a GIF/screenshot here -->
<!-- ![rikugan demo](assets/demo.gif) -->

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap onikukiraii/rikugan
brew install rikugan
```

### Go install

```bash
go install github.com/onikukiraii/rikugan@latest
```

### Build from source

```bash
git clone https://github.com/onikukiraii/rikugan.git
cd rikugan
go build -o rikugan .
```

## Usage

```bash
rikugan                     # unstaged changes
rikugan --staged            # staged changes
rikugan HEAD~3              # last 3 commits
rikugan main..feature       # branch diff
rikugan --version           # show version
```

## Workflow

1. Run `rikugan` in a git repository
2. Browse the diff — use `Tab` / `Shift+Tab` to switch files, or `Space` to fuzzy-search
3. Press `c` to add review comments on specific lines
4. Press `y` to copy the annotated diff to clipboard (or `Y` for comments only)
5. Paste into ChatGPT, Claude, or any AI assistant for code review

The clipboard output is structured as Markdown with inline `# >> COMMENT:` markers inside diff blocks, plus a summary section — optimized for LLM consumption.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `Ctrl+d` / `Ctrl+u` | Half page down / up |
| `gg` / `G` | Go to top / bottom |
| `]` / `[` | Next / previous hunk |
| `Tab` / `Shift+Tab` | Next / previous file |
| `Space` | Fuzzy file picker |

### View

| Key | Action |
|-----|--------|
| `V` | Toggle inline / split view |
| `h` / `l` | Switch pane (split mode) |
| `?` | Toggle help |

### Comments & Clipboard

| Key | Action |
|-----|--------|
| `c` | Add comment to current line |
| `d` | Delete comment on current line |
| `y` | Copy diff + comments to clipboard |
| `Y` | Copy comments only to clipboard |

### General

| Key | Action |
|-----|--------|
| `q` / `Esc` | Quit |

## Requirements

- Git
- A terminal with true-color support (for syntax highlighting)

## Tech Stack

- [Go](https://go.dev/)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Chroma](https://github.com/alecthomas/chroma) — Syntax highlighting

## Contributing

Contributions are welcome! Feel free to open issues and pull requests.

```bash
# Run tests
go test ./...

# Build
go build -o rikugan .
```

## License

[MIT](LICENSE)
