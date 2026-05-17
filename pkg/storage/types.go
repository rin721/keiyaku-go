package storage

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
)

// FileSystemType identifies the underlying storage implementation.
type FileSystemType string

// Config describes a Storage instance.
type Config struct {
	Type          FileSystemType
	Root          string
	FS            afero.Fs
	FilePerm      os.FileMode
	DirPerm       os.FileMode
	WatchDebounce time.Duration
}

// Option adjusts Storage configuration for constructor helpers.
type Option func(*Config)

// Storage is a concurrency-safe facade over file operations.
type Storage struct {
	fs            afero.Fs
	fsType        FileSystemType
	root          string
	filePerm      os.FileMode
	dirPerm       os.FileMode
	watchDebounce time.Duration
	osRooted      bool
	readOnly      bool
	mu            sync.RWMutex
}

// MIMEInfo describes a detected content type.
type MIMEInfo struct {
	MIME      string
	Extension string
}

// CopyOptions describes copy behavior.
type CopyOptions struct {
	Overwrite bool
}

// CopyOption adjusts copy behavior.
type CopyOption func(*CopyOptions)

// WatchOp is the normalized file watcher operation.
type WatchOp string

// Event describes a normalized file system event.
type Event struct {
	Path string
	Op   WatchOp
	Raw  fsnotify.Op
	Time time.Time
}

// WatchOptions describes watcher behavior.
type WatchOptions struct {
	Recursive bool
	Debounce  time.Duration
	Buffer    int
}

// WatchOption adjusts watcher behavior.
type WatchOption func(*WatchOptions)

// Watcher exposes normalized file events and watcher errors.
type Watcher struct {
	events   chan Event
	errors   chan error
	raw      *fsnotify.Watcher
	done     chan struct{}
	closeErr error
	once     sync.Once
}

// Type returns the configured file system type.
func (s *Storage) Type() FileSystemType {
	if s == nil {
		return ""
	}
	return s.fsType
}

// Root returns the configured OS root for OS-backed storage.
func (s *Storage) Root() string {
	if s == nil {
		return ""
	}
	return s.root
}

// Events returns normalized file system events.
func (w *Watcher) Events() <-chan Event {
	if w == nil {
		return nil
	}
	return w.events
}

// Errors returns watcher errors.
func (w *Watcher) Errors() <-chan error {
	if w == nil {
		return nil
	}
	return w.errors
}
