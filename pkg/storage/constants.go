package storage

import "time"

const (
	// FileSystemOS uses the host operating system under a configured root.
	FileSystemOS FileSystemType = "os"
	// FileSystemMemory uses an in-memory afero file system.
	FileSystemMemory FileSystemType = "memory"
	// FileSystemReadOnly wraps another storage as read-only.
	FileSystemReadOnly FileSystemType = "readonly"
	// FileSystemCustom uses a caller-supplied afero file system.
	FileSystemCustom FileSystemType = "custom"
)

const (
	defaultFilePerm      = 0o644
	defaultDirPerm       = 0o755
	defaultWatchDebounce = 100 * time.Millisecond
	defaultWatchBuffer   = 16
)

const (
	// WatchOpCreate represents file or directory creation.
	WatchOpCreate WatchOp = "create"
	// WatchOpWrite represents file content or metadata writes.
	WatchOpWrite WatchOp = "write"
	// WatchOpRemove represents file or directory removal.
	WatchOpRemove WatchOp = "remove"
	// WatchOpRename represents file or directory rename.
	WatchOpRename WatchOp = "rename"
	// WatchOpChmod represents permission changes.
	WatchOpChmod WatchOp = "chmod"
)
