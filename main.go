package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/onikukiraii/rikugan/internal/diff"
	"github.com/onikukiraii/rikugan/internal/tui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "show version")
	staged := flag.Bool("staged", false, "show staged changes (--cached)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "rikugan - TUI diff reviewer for AI prompts\n\n")
		fmt.Fprintf(os.Stderr, "Usage: rikugan [flags] [git-diff-args...]\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  rikugan                    # unstaged changes\n")
		fmt.Fprintf(os.Stderr, "  rikugan --staged            # staged changes\n")
		fmt.Fprintf(os.Stderr, "  rikugan HEAD~3              # last 3 commits\n")
		fmt.Fprintf(os.Stderr, "  rikugan main..feature       # branch diff\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("rikugan %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if *staged {
		args = append([]string{"--cached"}, args...)
	}

	files, err := diff.Run(args)
	if err != nil {
		runWithError(err)
		return
	}

	m := tui.New(files)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runWithError(err error) {
	m := tui.NewError(err)
	p := tea.NewProgram(m)
	if _, runErr := p.Run(); runErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", runErr)
		os.Exit(1)
	}
}
