package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	content := "See [Installation](01-installation.md) and [broken](nope.md) and [ext](https://example.com) and [anchor](#section)."
	links := extractLinks(content)
	want := []string{"01-installation.md", "nope.md", "https://example.com", "#section"}
	if len(links) != len(want) {
		t.Fatalf("got %d links, want %d: %v", len(links), len(want), links)
	}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("link %d: got %q, want %q", i, links[i], w)
		}
	}
}

func TestCheckFileBrokenLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target.md"), []byte("# ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "source.md")
	body := "[good](target.md) and [bad](missing.md) and [ext](https://example.com)"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	broken, err := checkFile(src)
	if err != nil {
		t.Fatalf("checkFile: %v", err)
	}
	if len(broken) != 1 || broken[0] != "missing.md" {
		t.Fatalf("got broken links %v, want [missing.md]", broken)
	}
}
