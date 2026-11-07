package pathguard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var AllowedRoots = []string{
	".cache",
	".npm",
	"Library/Caches",
	"Library/Application Support",
	"Library/Logs",
	"Library/pnpm/store",
	"Library/Developer/Xcode/DerivedData",
	"Library/Developer/CoreSimulator/Caches",
}

type UnsafePathError struct {
	Path   string
	Reason string
}

func (e *UnsafePathError) Error() string {
	return fmt.Sprintf("refusing to touch %s: %s", e.Path, e.Reason)
}

var ErrUnsafePath = errors.New("unsafe path")

func (e *UnsafePathError) Is(target error) bool {
	return target == ErrUnsafePath
}

type Guard struct {
	roots []string
}

func New(home string) *Guard {
	roots := make([]string, len(AllowedRoots))
	for i, root := range AllowedRoots {
		roots[i] = filepath.Join(home, root)
	}
	return &Guard{roots: roots}
}

func IsUnderAllowedRoot(pattern string) bool {
	for _, root := range AllowedRoots {
		if strings.HasPrefix(pattern, "~/"+root+"/") {
			return true
		}
	}
	return false
}

func isInside(path, root string) bool {
	return strings.HasPrefix(path, root+string(filepath.Separator))
}

func (g *Guard) rootFor(path string) (string, bool) {
	for _, root := range g.roots {
		if isInside(path, root) {
			return root, true
		}
	}
	return "", false
}

func (g *Guard) Check(path string) string {
	if !filepath.IsAbs(path) {
		return "path is not absolute"
	}
	if filepath.Clean(path) != path {
		return "path is not normalized"
	}

	root, ok := g.rootFor(path)
	if !ok {
		return "path is outside the allowed cache folders"
	}

	info, err := os.Lstat(path)
	if err != nil {
		return "path does not exist"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "path is a symbolic link"
	}

	realParent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return "could not resolve the real location"
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "could not resolve the real location"
	}

	if !isInside(filepath.Join(realParent, filepath.Base(path)), realRoot) {
		return "path resolves outside the allowed cache folders"
	}

	return ""
}

func (g *Guard) AssertSafe(path string) error {
	if reason := g.Check(path); reason != "" {
		return &UnsafePathError{Path: path, Reason: reason}
	}
	return nil
}
