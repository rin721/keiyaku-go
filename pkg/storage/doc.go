// Package storage provides a small, business-agnostic file processing facade.
//
// The package wraps afero-backed file systems, file copying, MIME detection and
// OS-backed file watching behind repository-local types. Business code should
// compose Storage from cmd or internal adapters instead of importing concrete
// file system libraries directly.
package storage
