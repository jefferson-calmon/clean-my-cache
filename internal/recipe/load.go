package recipe

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"sort"

	"go.yaml.in/yaml/v3"
)

const electronSupportRoot = "~/Library/Application Support/"

var electronCacheFolders = []string{"Cache", "Code Cache", "GPUCache", "DawnGraphiteCache", "DawnWebGPUCache"}

func applyPreset(r Recipe) Recipe {
	if r.Preset != "electron" {
		return r
	}

	base := electronSupportRoot + r.SupportFolder
	paths := make([]string, 0, len(electronCacheFolders)+len(r.Paths))
	for _, folder := range electronCacheFolders {
		paths = append(paths, base+"/"+folder)
	}
	r.Paths = append(paths, r.Paths...)
	return r
}

func decode(data []byte, source string) (Recipe, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var r Recipe
	if err := decoder.Decode(&r); err != nil {
		return Recipe{}, fmt.Errorf("%s: %w", source, err)
	}

	r.Source = source
	r = applyPreset(r)
	if err := Validate(r); err != nil {
		return Recipe{}, fmt.Errorf("%s: %w", source, err)
	}
	return r, nil
}

func Load(fsys fs.FS, sourcePrefix string) (Set, error) {
	files, err := fs.Glob(fsys, "*/*.yaml")
	if err != nil {
		return nil, err
	}
	topLevel, err := fs.Glob(fsys, "*.yaml")
	if err != nil {
		return nil, err
	}
	files = append(files, topLevel...)

	var set Set
	var errs []error
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		r, err := decode(data, path.Join(sourcePrefix, file))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		set = append(set, r)
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return sorted(set), nil
}

func LoadDir(dir string) (Set, error) {
	info, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !info.IsDir()) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return Load(os.DirFS(dir), dir)
}

func Merge(sets ...Set) (Set, error) {
	seen := map[string]string{}
	var merged Set
	var errs []error
	for _, set := range sets {
		for _, r := range set {
			if previous, ok := seen[r.ID]; ok {
				errs = append(errs, fmt.Errorf("duplicate recipe id %q in %s and %s", r.ID, previous, r.Source))
				continue
			}
			seen[r.ID] = r.Source
			merged = append(merged, r)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return sorted(merged), nil
}

func sorted(set Set) Set {
	out := slices.Clone(set)
	sort.SliceStable(out, func(i, j int) bool {
		ci, cj := slices.Index(Categories, out[i].Category), slices.Index(Categories, out[j].Category)
		if ci != cj {
			return ci < cj
		}
		return out[i].ID < out[j].ID
	})
	return out
}
