package clean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jefferson-calmon/clean-my-cache/internal/pathguard"
	"github.com/jefferson-calmon/clean-my-cache/internal/recipe"
	"github.com/jefferson-calmon/clean-my-cache/internal/scan"
	"github.com/jefferson-calmon/clean-my-cache/internal/size"
	"github.com/jefferson-calmon/clean-my-cache/internal/system"
)

type Mode int

const (
	ModeDelete Mode = iota
	ModeTrash
)

type Outcome struct {
	Item       scan.Item
	OK         bool
	FreedBytes int64
	Err        error
}

type Cleaner struct {
	Home   string
	Runner system.Runner
	Guard  *pathguard.Guard
	Mode   Mode
	Now    func() time.Time
}

func (c *Cleaner) moveToTrash(path string) error {
	trash := filepath.Join(c.Home, ".Trash")
	if err := os.MkdirAll(trash, 0o700); err != nil {
		return err
	}
	stamp := c.Now().Format("2006-01-02 15.04.05.000")
	return os.Rename(path, filepath.Join(trash, fmt.Sprintf("%s %s", filepath.Base(path), stamp)))
}

func (c *Cleaner) remove(path string) error {
	if err := c.Guard.AssertSafe(path); err != nil {
		return err
	}
	if c.Mode == ModeTrash {
		return c.moveToTrash(path)
	}
	return os.RemoveAll(path)
}

func (c *Cleaner) cleanPaths(item scan.Item) Outcome {
	var errs []error
	for _, path := range item.Paths {
		if err := c.remove(path); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return Outcome{Item: item, OK: true, FreedBytes: item.Bytes}
	}
	freed := max(item.Bytes-size.Of(item.Paths), 0)
	return Outcome{Item: item, FreedBytes: freed, Err: errors.Join(errs...)}
}

func (c *Cleaner) cleanCommand(ctx context.Context, item scan.Item) Outcome {
	r := item.Recipe
	if !c.Runner.Run(ctx, r.Run).OK {
		return Outcome{Item: item, Err: fmt.Errorf("%s failed", strings.Join(r.Run, " "))}
	}
	if !item.Measured {
		return Outcome{Item: item, OK: true}
	}
	return Outcome{Item: item, OK: true, FreedBytes: max(item.Bytes-size.Of(item.Paths), 0)}
}

func (c *Cleaner) Clean(ctx context.Context, item scan.Item) Outcome {
	if item.Recipe.Kind() == recipe.KindCommand {
		return c.cleanCommand(ctx, item)
	}
	return c.cleanPaths(item)
}
