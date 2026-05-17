package storage

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

// WithFilePerm sets the default permission used when creating files.
func WithFilePerm(perm os.FileMode) Option {
	return func(config *Config) {
		config.FilePerm = perm
	}
}

// WithDirPerm sets the default permission used when creating directories.
func WithDirPerm(perm os.FileMode) Option {
	return func(config *Config) {
		config.DirPerm = perm
	}
}

// WithWatchDebounce sets the default watcher debounce duration.
func WithWatchDebounce(duration time.Duration) Option {
	return func(config *Config) {
		config.WatchDebounce = duration
	}
}

// WithAferoFS sets a custom afero file system for New.
func WithAferoFS(fs afero.Fs) Option {
	return func(config *Config) {
		config.FS = fs
	}
}

// WithOverwrite allows Copy to replace an existing target.
func WithOverwrite(overwrite bool) CopyOption {
	return func(options *CopyOptions) {
		options.Overwrite = overwrite
	}
}

// WithRecursiveWatch enables recursive directory watching.
func WithRecursiveWatch(recursive bool) WatchOption {
	return func(options *WatchOptions) {
		options.Recursive = recursive
	}
}

// WithEventDebounce overrides the watcher debounce duration for one watcher.
func WithEventDebounce(duration time.Duration) WatchOption {
	return func(options *WatchOptions) {
		options.Debounce = duration
	}
}

// WithWatchBuffer sets the watcher event and error channel buffer size.
func WithWatchBuffer(size int) WatchOption {
	return func(options *WatchOptions) {
		options.Buffer = size
	}
}

func applyOptions(config Config, opts []Option) Config {
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}
	return withDefaults(config)
}

func withDefaults(config Config) Config {
	if config.FilePerm == 0 {
		config.FilePerm = defaultFilePerm
	}
	if config.DirPerm == 0 {
		config.DirPerm = defaultDirPerm
	}
	if config.WatchDebounce == 0 {
		config.WatchDebounce = defaultWatchDebounce
	}
	return config
}

func defaultCopyOptions(opts []CopyOption) CopyOptions {
	options := CopyOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func defaultWatchOptions(debounce time.Duration, opts []WatchOption) WatchOptions {
	options := WatchOptions{
		Debounce: debounce,
		Buffer:   defaultWatchBuffer,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if options.Buffer <= 0 {
		options.Buffer = defaultWatchBuffer
	}
	return options
}
