// Command verifydist checks that the browser distribution is self-contained.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	htmlReference  = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["'](?:\./)?([^"'?#]+)`)
	fetchReference = regexp.MustCompile(`(?i)fetch\(\s*["'](?:\./)?([^"'?#]+)`)
	gameReference  = regexp.MustCompile(`assets/(?:images|audio)/[a-z0-9][a-z0-9._-]*`)
)

func main() {
	repository := flag.String("repo", ".", "repository root")
	distribution := flag.String("dist", "dist", "distribution directory, relative to repository root")
	flag.Parse()
	if err := verify(*repository, filepath.Join(*repository, *distribution)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func verify(repository, distribution string) error {
	indexPath := filepath.Join(distribution, "index.html")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read browser shell: %w", err)
	}
	if !strings.Contains(string(index), `id="gameCanvas"`) {
		return errors.New("index.html does not contain required gameCanvas")
	}

	references := parseReferences(index)
	sourceRoot := filepath.Join(repository, "internal", "platform")
	err = filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, match := range gameReference.FindAllString(string(contents), -1) {
			references[match] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scan game asset references: %w", err)
	}

	paths := make([]string, 0, len(references))
	for path := range references {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		info, statErr := os.Stat(filepath.Join(distribution, filepath.FromSlash(path)))
		if statErr != nil {
			return fmt.Errorf("referenced file %q is missing: %w", path, statErr)
		}
		if info.IsDir() || info.Size() == 0 {
			return fmt.Errorf("referenced file %q is empty or not a file", path)
		}
	}
	if _, err = os.Stat(filepath.Join(distribution, ".nojekyll")); err != nil {
		return fmt.Errorf("distribution is missing .nojekyll: %w", err)
	}
	fmt.Printf("verified %d browser and game references in %s\n", len(paths), distribution)
	return nil
}

func parseReferences(index []byte) map[string]struct{} {
	references := make(map[string]struct{})
	for _, expression := range []*regexp.Regexp{htmlReference, fetchReference} {
		for _, match := range expression.FindAllSubmatch(index, -1) {
			path := string(match[1])
			if strings.Contains(path, "://") || strings.HasPrefix(path, "data:") {
				continue
			}
			references[path] = struct{}{}
		}
	}
	return references
}
