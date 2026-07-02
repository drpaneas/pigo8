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

func TestStripCodeFences(t *testing.T) {
	content := "Real [link](target.md) here.\n```go\nfunc Foo[T any](x T) { return }\n// [fake](not-a-real-link)\n```\nAnother [real](other.md) link."
	stripped := stripCodeFences(content)
	links := extractLinks(stripped)
	want := []string{"target.md", "other.md"}
	if len(links) != len(want) {
		t.Fatalf("got links %v, want %v", links, want)
	}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("link %d: got %q, want %q", i, links[i], w)
		}
	}
}

func TestStripInlineCode(t *testing.T) {
	content := "Real [link](target.md) here. Example syntax `[text](target)` is not a real link. Another [real](other.md) link."
	stripped := stripInlineCode(content)
	links := extractLinks(stripped)
	want := []string{"target.md", "other.md"}
	if len(links) != len(want) {
		t.Fatalf("got links %v, want %v", links, want)
	}
	for i, w := range want {
		if links[i] != w {
			t.Errorf("link %d: got %q, want %q", i, links[i], w)
		}
	}
}

func TestIsCheckable(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"empty", "", false},
		{"anchor", "#section", false},
		{"external https", "https://example.com", false},
		{"external http", "http://example.com", false},
		{"mailto", "mailto:foo@example.com", false},
		{"relative path", "01-installation.md", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCheckable(tt.target); got != tt.want {
				t.Errorf("isCheckable(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestCheckFileAnchorFragment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "page.md"), []byte("# ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "source.md")
	body := "[section link](page.md#section)"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	broken, err := checkFile(src)
	if err != nil {
		t.Fatalf("checkFile: %v", err)
	}
	if len(broken) != 0 {
		t.Fatalf("expected no broken links (anchor should resolve to existing page.md), got %v", broken)
	}
}

func TestCheckFileLinkWithTitle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target.md"), []byte("# ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "source.md")
	body := `[text](target.md "My Title")`
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	broken, err := checkFile(src)
	if err != nil {
		t.Fatalf("checkFile: %v", err)
	}
	if len(broken) != 0 {
		t.Fatalf("expected no broken links (title syntax should resolve to target.md), got %v", broken)
	}
}
