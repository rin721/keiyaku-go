package storage

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// New builds a Storage from Config.
func New(config Config) (*Storage, error) {
	config = withDefaults(config)

	switch config.Type {
	case FileSystemOS:
		return newOS(config)
	case FileSystemMemory:
		return newMemory(config), nil
	case FileSystemReadOnly:
		if config.FS == nil {
			return nil, invalidConfig("readonly storage requires an afero fs")
		}
		return &Storage{
			fs:            afero.NewReadOnlyFs(config.FS),
			fsType:        FileSystemReadOnly,
			root:          config.Root,
			filePerm:      config.FilePerm,
			dirPerm:       config.DirPerm,
			watchDebounce: config.WatchDebounce,
			readOnly:      true,
		}, nil
	case FileSystemCustom:
		if config.FS == nil {
			return nil, invalidConfig("custom storage requires an afero fs")
		}
		return &Storage{
			fs:            config.FS,
			fsType:        FileSystemCustom,
			root:          config.Root,
			filePerm:      config.FilePerm,
			dirPerm:       config.DirPerm,
			watchDebounce: config.WatchDebounce,
		}, nil
	default:
		return nil, invalidConfig("unsupported file system type %q", config.Type)
	}
}

// NewOS builds an OS-backed Storage rooted at root.
func NewOS(root string, opts ...Option) (*Storage, error) {
	config := applyOptions(Config{Type: FileSystemOS, Root: root}, opts)
	return New(config)
}

// NewMemory builds an in-memory Storage.
func NewMemory(opts ...Option) (*Storage, error) {
	config := applyOptions(Config{Type: FileSystemMemory}, opts)
	return New(config)
}

// NewReadOnly wraps another Storage with a read-only file system.
func NewReadOnly(base *Storage, opts ...Option) (*Storage, error) {
	if base == nil {
		return nil, invalidConfig("readonly storage requires a base storage")
	}
	config := applyOptions(Config{
		Type:          FileSystemReadOnly,
		Root:          base.root,
		FS:            base.fs,
		FilePerm:      base.filePerm,
		DirPerm:       base.dirPerm,
		WatchDebounce: base.watchDebounce,
	}, opts)
	storage, err := New(config)
	if err != nil {
		return nil, err
	}
	storage.osRooted = base.osRooted
	storage.readOnly = true
	return storage, nil
}

func newOS(config Config) (*Storage, error) {
	if config.Root == "" {
		return nil, invalidConfig("os storage root is required")
	}
	root, err := filepath.Abs(config.Root)
	if err != nil {
		return nil, invalidConfig("resolve os root %q: %v", config.Root, err)
	}
	if err := os.MkdirAll(root, config.DirPerm); err != nil {
		return nil, invalidConfig("create os root %q: %v", root, err)
	}
	return &Storage{
		fs:            afero.NewBasePathFs(afero.NewOsFs(), root),
		fsType:        FileSystemOS,
		root:          root,
		filePerm:      config.FilePerm,
		dirPerm:       config.DirPerm,
		watchDebounce: config.WatchDebounce,
		osRooted:      true,
	}, nil
}

func newMemory(config Config) *Storage {
	return &Storage{
		fs:            afero.NewMemMapFs(),
		fsType:        FileSystemMemory,
		filePerm:      config.FilePerm,
		dirPerm:       config.DirPerm,
		watchDebounce: config.WatchDebounce,
	}
}

// Exists reports whether name exists.
func (s *Storage) Exists(name string) (bool, error) {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.existsLocked(cleaned)
}

// Stat returns file information for name.
func (s *Storage) Stat(name string) (os.FileInfo, error) {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, err := s.fs.Stat(cleaned)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// ReadFile reads the complete file at name.
func (s *Storage) ReadFile(name string) ([]byte, error) {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return afero.ReadFile(s.fs, cleaned)
}

// WriteFile writes data to name, creating parent directories when needed.
func (s *Storage) WriteFile(name string, data []byte) error {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return err
	}
	if err := rejectRootMutation(cleaned); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureParentLocked(cleaned); err != nil {
		return err
	}
	return afero.WriteFile(s.fs, cleaned, data, s.filePerm)
}

// MkdirAll creates name and any missing parent directories.
func (s *Storage) MkdirAll(name string) error {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fs.MkdirAll(cleaned, s.dirPerm)
}

// Remove removes name.
func (s *Storage) Remove(name string) error {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return err
	}
	if err := rejectRootMutation(cleaned); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fs.Remove(cleaned)
}

// RemoveAll removes name and any children.
func (s *Storage) RemoveAll(name string) error {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return err
	}
	if err := rejectRootMutation(cleaned); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fs.RemoveAll(cleaned)
}

// Rename renames oldName to newName, creating the target parent when needed.
func (s *Storage) Rename(oldName string, newName string) error {
	oldCleaned, err := cleanStoragePath(oldName)
	if err != nil {
		return err
	}
	newCleaned, err := cleanStoragePath(newName)
	if err != nil {
		return err
	}
	if err := rejectRootMutation(oldCleaned); err != nil {
		return err
	}
	if err := rejectRootMutation(newCleaned); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureParentLocked(newCleaned); err != nil {
		return err
	}
	return s.fs.Rename(oldCleaned, newCleaned)
}

// Walk walks the tree rooted at root.
func (s *Storage) Walk(root string, fn filepath.WalkFunc) error {
	if fn == nil {
		return invalidConfig("walk callback is required")
	}
	cleaned, err := cleanStoragePath(root)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return afero.Walk(s.fs, cleaned, func(name string, info os.FileInfo, walkErr error) error {
		return fn(publicPath(name), info, walkErr)
	})
}

func (s *Storage) existsLocked(name string) (bool, error) {
	_, err := s.fs.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Storage) ensureParentLocked(name string) error {
	parent := filepath.Dir(name)
	if parent == "." {
		return nil
	}
	return s.fs.MkdirAll(parent, s.dirPerm)
}

func (s *Storage) realPath(name string) string {
	if name == "." {
		return s.root
	}
	return filepath.Join(s.root, name)
}
