package size

import (
	"io/fs"
	"path/filepath"
	"sync"
	"syscall"
)

const blockSize = 512

type inode struct {
	dev uint64
	ino uint64
}

type counter struct {
	mu    sync.Mutex
	seen  map[inode]bool
	total int64
}

func (c *counter) add(info fs.FileInfo) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		c.mu.Lock()
		c.total += info.Size()
		c.mu.Unlock()
		return
	}

	key := inode{dev: uint64(stat.Dev), ino: stat.Ino}
	c.mu.Lock()
	defer c.mu.Unlock()
	if stat.Nlink > 1 {
		if c.seen[key] {
			return
		}
		c.seen[key] = true
	}
	c.total += stat.Blocks * blockSize
}

func walk(root string, c *counter) {
	_ = filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		c.add(info)
		return nil
	})
}

func Of(paths []string) int64 {
	c := &counter{seen: map[inode]bool{}}
	var wg sync.WaitGroup
	for _, path := range paths {
		wg.Add(1)
		go func() {
			defer wg.Done()
			walk(path, c)
		}()
	}
	wg.Wait()
	return c.total
}
