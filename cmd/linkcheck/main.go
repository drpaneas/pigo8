// Package main implements a documentation link checker for the PIGO8 mdBook site.
// It walks docs/ for markdown files, extracts relative links and image references,
// and reports any that don't resolve to a real file.
//
// Usage:
//
//	go run ./cmd/linkcheck -dir docs
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var linkPattern = regexp.MustCompile(`\]\(([^)]+)\)`)

// extractLinks returns every link target found inside markdown link/image
// syntax `[text](target)` or `![alt](target)`, in the order they appear.
func extractLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		links = append(links, strings.TrimSpace(m[1]))
	}
	return links
}

// isCheckable reports whether a link target should be verified against the
// filesystem (relative paths only; external URLs and in-page anchors are skipped).
func isCheckable(target string) bool {
	if target == "" {
		return false
	}
	if strings.HasPrefix(target, "#") {
		return false
	}
	if strings.Contains(target, "://") {
		return false
	}
	if strings.HasPrefix(target, "mailto:") {
		return false
	}
	return true
}

// checkFile reads the markdown file at path and returns any link targets that
// don't resolve to an existing file relative to the file's directory. Anchor
// fragments (e.g. "page.md#section") are stripped before resolution.
func checkFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	var broken []string
	for _, link := range extractLinks(string(data)) {
		if !isCheckable(link) {
			continue
		}
		target := link
		if idx := strings.Index(target, "#"); idx >= 0 {
			target = target[:idx]
		}
		if target == "" {
			continue
		}
		full := filepath.Join(dir, target)
		if _, err := os.Stat(full); err != nil {
			broken = append(broken, link)
		}
	}
	return broken, nil
}

func main() {
	dir := flag.String("dir", "docs", "Root directory of markdown files to check")
	flag.Parse()

	var totalBroken int
	err := filepath.WalkDir(*dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		broken, err := checkFile(path)
		if err != nil {
			return err
		}
		for _, b := range broken {
			fmt.Printf("%s: broken link -> %s\n", path, b)
			totalBroken++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking %s: %v\n", *dir, err)
		os.Exit(1)
	}

	if totalBroken > 0 {
		fmt.Printf("\n%d broken link(s) found.\n", totalBroken)
		os.Exit(1)
	}
	fmt.Println("No broken links found.")
}
