package tui

import (
	"strings"
)

// DiffSegment represents a segment of text that may or may not have changed.
type DiffSegment struct {
	Text    string
	Changed bool
}

// ComputeWordDiff computes character-level differences between old and new strings.
// Returns segments for old and new respectively.
func ComputeWordDiff(old, new string) ([]DiffSegment, []DiffSegment) {
	ops := editScript(old, new)

	var oldSegs, newSegs []DiffSegment
	var oldUnchanged, oldChanged strings.Builder
	var newUnchanged, newChanged strings.Builder

	flushOld := func() {
		if oldUnchanged.Len() > 0 {
			oldSegs = append(oldSegs, DiffSegment{Text: oldUnchanged.String(), Changed: false})
			oldUnchanged.Reset()
		}
		if oldChanged.Len() > 0 {
			oldSegs = append(oldSegs, DiffSegment{Text: oldChanged.String(), Changed: true})
			oldChanged.Reset()
		}
	}

	flushNew := func() {
		if newUnchanged.Len() > 0 {
			newSegs = append(newSegs, DiffSegment{Text: newUnchanged.String(), Changed: false})
			newUnchanged.Reset()
		}
		if newChanged.Len() > 0 {
			newSegs = append(newSegs, DiffSegment{Text: newChanged.String(), Changed: true})
			newChanged.Reset()
		}
	}

	oi, ni := 0, 0
	for _, op := range ops {
		switch op {
		case opEqual:
			if oldChanged.Len() > 0 || newChanged.Len() > 0 {
				flushOld()
				flushNew()
			}
			oldUnchanged.WriteByte(old[oi])
			newUnchanged.WriteByte(new[ni])
			oi++
			ni++
		case opDelete:
			if oldUnchanged.Len() > 0 {
				flushOld()
			}
			oldChanged.WriteByte(old[oi])
			oi++
		case opInsert:
			if newUnchanged.Len() > 0 {
				flushNew()
			}
			newChanged.WriteByte(new[ni])
			ni++
		}
	}
	flushOld()
	flushNew()

	return oldSegs, newSegs
}

type editOp int

const (
	opEqual editOp = iota
	opDelete
	opInsert
)

// editScript computes the Myers diff edit script between two strings.
func editScript(a, b string) []editOp {
	n, m := len(a), len(b)
	if n == 0 && m == 0 {
		return nil
	}
	if n == 0 {
		ops := make([]editOp, m)
		for i := range ops {
			ops[i] = opInsert
		}
		return ops
	}
	if m == 0 {
		ops := make([]editOp, n)
		for i := range ops {
			ops[i] = opDelete
		}
		return ops
	}

	// Simple DP-based LCS for reasonable line lengths
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to produce edit script
	var ops []editOp
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			ops = append(ops, opEqual)
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			ops = append(ops, opInsert)
			j--
		} else {
			ops = append(ops, opDelete)
			i--
		}
	}

	// Reverse
	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}
	return ops
}
