package pathguard

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandHome(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func expandSegment(bases []string, segment string) []string {
	if !strings.Contains(segment, "*") {
		out := make([]string, len(bases))
		for i, base := range bases {
			out[i] = filepath.Join(base, segment)
		}
		return out
	}

	var out []string
	for _, base := range bases {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if matched, _ := filepath.Match(segment, entry.Name()); matched {
				out = append(out, filepath.Join(base, entry.Name()))
			}
		}
	}
	return out
}

func ExpandGlob(pattern string) []string {
	if !strings.Contains(pattern, "*") {
		if exists(pattern) {
			return []string{pattern}
		}
		return nil
	}

	current := []string{string(filepath.Separator)}
	for _, segment := range strings.Split(pattern, string(filepath.Separator)) {
		if segment == "" {
			continue
		}
		current = expandSegment(current, segment)
		if len(current) == 0 {
			return nil
		}
	}

	var present []string
	for _, path := range current {
		if exists(path) {
			present = append(present, path)
		}
	}
	return present
}

func ExpandAll(patterns []string, home string) []string {
	seen := map[string]bool{}
	var out []string
	for _, pattern := range patterns {
		for _, path := range ExpandGlob(ExpandHome(pattern, home)) {
			if seen[path] {
				continue
			}
			seen[path] = true
			out = append(out, path)
		}
	}
	return out
}
