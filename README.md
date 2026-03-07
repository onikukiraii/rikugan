# rikugan

TUI diff reviewer for AI prompts. Review git diffs interactively, add inline comments, and copy everything to clipboard in a format ready for AI code review.

## Install

```bash
brew tap onikukiraii/rikugan
brew install rikugan
```

Or build from source:

```bash
go install github.com/onikukiraii/rikugan@latest
```

## Usage

```bash
rikugan                     # unstaged changes
rikugan --staged            # staged changes
rikugan HEAD~3              # last 3 commits
rikugan main..feature       # branch diff
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `Ctrl+d` / `Ctrl+u` | Half page down / up |
| `gg` / `G` | Go to top / bottom |
| `Tab` / `Shift+Tab` | Next / previous file |
| `V` | Toggle inline / split view |
| `h` / `l` | Switch pane (split mode) |
| `c` | Add comment to current line |
| `d` | Delete comment on current line |
| `y` | Copy diff + comments to clipboard |
| `?` | Toggle help |
| `q` | Quit |

## How it works

1. Run `rikugan` in a git repository
2. Navigate the diff with vim-style keybindings
3. Press `c` to add review comments on specific lines
4. Press `y` to copy the annotated diff to clipboard
5. Paste into your AI assistant for code review

## License

MIT
