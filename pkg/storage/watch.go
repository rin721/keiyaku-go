package storage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch starts watching path for file system changes.
func (s *Storage) Watch(name string, opts ...WatchOption) (*Watcher, error) {
	cleaned, err := cleanStoragePath(name)
	if err != nil {
		return nil, err
	}
	if !s.osRooted {
		return nil, unsupported("watch requires an os-rooted storage")
	}
	options := defaultWatchOptions(s.watchDebounce, opts)
	raw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	realName := s.realPath(cleaned)
	if options.Recursive {
		err = addRecursive(raw, realName)
	} else {
		err = raw.Add(realName)
	}
	if err != nil {
		_ = raw.Close()
		return nil, err
	}

	watcher := &Watcher{
		events: make(chan Event, options.Buffer),
		errors: make(chan error, options.Buffer),
		raw:    raw,
		done:   make(chan struct{}),
	}
	go watcher.run(s.root, options)
	return watcher, nil
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	if w == nil {
		return nil
	}
	w.once.Do(func() {
		close(w.done)
		w.closeErr = w.raw.Close()
	})
	return w.closeErr
}

func (w *Watcher) run(root string, options WatchOptions) {
	defer close(w.events)
	defer close(w.errors)

	pending := make(map[string]Event)
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}

	flush := func() {
		if len(pending) == 0 {
			return
		}
		events := pending
		pending = make(map[string]Event)
		for _, event := range events {
			select {
			case w.events <- event:
			case <-w.done:
				return
			}
		}
	}

	for {
		select {
		case <-w.done:
			flush()
			return
		case event, ok := <-w.raw.Events:
			if !ok {
				flush()
				return
			}
			if options.Recursive && event.Op&fsnotify.Create != 0 {
				w.addCreatedDirectory(event.Name)
			}
			normalized, err := normalizeEvent(root, event)
			if err != nil {
				w.sendError(err)
				continue
			}
			if options.Debounce <= 0 {
				select {
				case w.events <- normalized:
				case <-w.done:
					return
				}
				continue
			}
			pending[normalized.Path+"|"+string(normalized.Op)] = normalized
			resetTimer(timer, options.Debounce)
		case err, ok := <-w.raw.Errors:
			if !ok {
				continue
			}
			w.sendError(err)
		case <-timer.C:
			flush()
		}
	}
}

func (w *Watcher) sendError(err error) {
	select {
	case w.errors <- err:
	case <-w.done:
	}
}

func (w *Watcher) addCreatedDirectory(name string) {
	info, err := os.Stat(name)
	if err != nil || !info.IsDir() {
		return
	}
	_ = addRecursive(w.raw, name)
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

func addRecursive(watcher *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return watcher.Add(root)
	}
	return filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		return watcher.Add(name)
	})
}

func normalizeEvent(root string, event fsnotify.Event) (Event, error) {
	rel, err := filepath.Rel(root, event.Name)
	if err != nil {
		return Event{}, err
	}
	return Event{
		Path: publicPath(rel),
		Op:   normalizeWatchOp(event.Op),
		Raw:  event.Op,
		Time: time.Now(),
	}, nil
}

func normalizeWatchOp(op fsnotify.Op) WatchOp {
	switch {
	case op&fsnotify.Create != 0:
		return WatchOpCreate
	case op&fsnotify.Write != 0:
		return WatchOpWrite
	case op&fsnotify.Remove != 0:
		return WatchOpRemove
	case op&fsnotify.Rename != 0:
		return WatchOpRename
	case op&fsnotify.Chmod != 0:
		return WatchOpChmod
	default:
		return WatchOp("")
	}
}
