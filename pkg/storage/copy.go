package storage

import (
	"os"
	"path/filepath"

	otiai "github.com/otiai10/copy"
	"github.com/spf13/afero"
)

// Copy copies a file or directory within the same Storage.
func (s *Storage) Copy(src string, dst string, opts ...CopyOption) error {
	srcCleaned, err := cleanStoragePath(src)
	if err != nil {
		return err
	}
	dstCleaned, err := cleanStoragePath(dst)
	if err != nil {
		return err
	}
	if err := rejectRootMutation(dstCleaned); err != nil {
		return err
	}
	options := defaultCopyOptions(opts)

	s.mu.Lock()
	defer s.mu.Unlock()

	if exists, err := s.existsLocked(dstCleaned); err != nil {
		return err
	} else if exists {
		if !options.Overwrite {
			return alreadyExists(dst)
		}
		if err := s.fs.RemoveAll(dstCleaned); err != nil {
			return err
		}
	}

	if s.osRooted && !s.readOnly {
		return s.copyOSLocked(srcCleaned, dstCleaned)
	}
	return s.copyAferoLocked(srcCleaned, dstCleaned)
}

func (s *Storage) copyOSLocked(src string, dst string) error {
	srcReal := s.realPath(src)
	dstReal := s.realPath(dst)
	if err := os.MkdirAll(filepath.Dir(dstReal), s.dirPerm); err != nil {
		return err
	}
	return otiai.Copy(srcReal, dstReal)
}

func (s *Storage) copyAferoLocked(src string, dst string) error {
	info, err := s.fs.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return s.copyFileLocked(src, dst, info.Mode().Perm())
	}

	return afero.Walk(s.fs, src, func(name string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, name)
		if err != nil {
			return err
		}
		target := dst
		if rel != "." {
			target = filepath.Join(dst, rel)
		}
		if info.IsDir() {
			return s.fs.MkdirAll(target, info.Mode().Perm())
		}
		return s.copyFileLocked(name, target, info.Mode().Perm())
	})
}

func (s *Storage) copyFileLocked(src string, dst string, perm os.FileMode) error {
	if perm == 0 {
		perm = s.filePerm
	}
	if err := s.ensureParentLocked(dst); err != nil {
		return err
	}
	data, err := afero.ReadFile(s.fs, src)
	if err != nil {
		return err
	}
	return afero.WriteFile(s.fs, dst, data, perm)
}
