package diff

import (
	"testing"
)

const sampleDiff = `diff --git a/hello.go b/hello.go
index 1234567..abcdefg 100644
--- a/hello.go
+++ b/hello.go
@@ -1,5 +1,6 @@
 package main

 func main() {
-	fmt.Println("hello")
+	fmt.Println("hello, world")
+	fmt.Println("goodbye")
 }
`

func TestParse(t *testing.T) {
	files, err := Parse(sampleDiff)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.DisplayName() != "hello.go" {
		t.Errorf("expected hello.go, got %s", f.DisplayName())
	}

	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}

	h := f.Hunks[0]
	if h.OldStart != 1 || h.OldCount != 5 {
		t.Errorf("old range: expected 1,5 got %d,%d", h.OldStart, h.OldCount)
	}
	if h.NewStart != 1 || h.NewCount != 6 {
		t.Errorf("new range: expected 1,6 got %d,%d", h.NewStart, h.NewCount)
	}

	// Count line types
	var added, removed, context int
	for _, line := range h.Lines {
		switch line.Type {
		case LineAdded:
			added++
		case LineRemoved:
			removed++
		case LineContext:
			context++
		}
	}
	if added != 2 {
		t.Errorf("expected 2 added lines, got %d", added)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed line, got %d", removed)
	}
	if context != 4 {
		t.Errorf("expected 4 context lines, got %d", context)
	}
}

func TestParseMultiFile(t *testing.T) {
	multiDiff := `diff --git a/a.go b/a.go
index 1111111..2222222 100644
--- a/a.go
+++ b/a.go
@@ -1,3 +1,3 @@
 package a

-var x = 1
+var x = 2
diff --git a/b.go b/b.go
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/b.go
@@ -0,0 +1,3 @@
+package b
+
+var y = 1
`
	files, err := Parse(multiDiff)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].DisplayName() != "a.go" {
		t.Errorf("file 0: expected a.go, got %s", files[0].DisplayName())
	}
	if files[1].DisplayName() != "b.go" {
		t.Errorf("file 1: expected b.go, got %s", files[1].DisplayName())
	}
}
