package storage

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidConfig means the storage configuration is incomplete or invalid.
	ErrInvalidConfig = errors.New("storage: invalid config")
	// ErrInvalidPath means a path is empty, absolute, or escapes the storage root.
	ErrInvalidPath = errors.New("storage: invalid path")
	// ErrUnsupported means the current storage backend cannot perform the operation.
	ErrUnsupported = errors.New("storage: unsupported operation")
	// ErrAlreadyExists means the target exists and overwrite was not enabled.
	ErrAlreadyExists = errors.New("storage: target already exists")
)

func invalidConfig(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, fmt.Sprintf(format, args...))
}

func invalidPath(name string) error {
	return fmt.Errorf("%w: %q", ErrInvalidPath, name)
}

func unsupported(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrUnsupported, fmt.Sprintf(format, args...))
}

func alreadyExists(name string) error {
	return fmt.Errorf("%w: %q", ErrAlreadyExists, name)
}
