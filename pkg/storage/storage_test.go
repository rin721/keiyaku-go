package storage

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMemoryStorageFileOperations(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	if err := store.WriteFile("dir/file.txt", []byte("hello")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	exists, err := store.Exists("dir/file.txt")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Fatal("file should exist")
	}
	content, err := store.ReadFile("dir/file.txt")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("content = %q", content)
	}
	if _, err := store.Stat("dir/file.txt"); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	var walked []string
	err = store.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		walked = append(walked, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}
	sort.Strings(walked)
	if !contains(walked, "dir/file.txt") {
		t.Fatalf("walked paths = %#v", walked)
	}

	if err := store.Rename("dir/file.txt", "renamed/file.txt"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if err := store.Remove("renamed/file.txt"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	exists, err = store.Exists("renamed/file.txt")
	if err != nil {
		t.Fatalf("Exists() after remove error = %v", err)
	}
	if exists {
		t.Fatal("file should not exist after remove")
	}
}

func TestStorageRejectsInvalidPaths(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	cases := []string{"", "../escape.txt", "/absolute.txt", `C:\absolute.txt`}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := store.ReadFile(name); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("ReadFile(%q) error = %v, want ErrInvalidPath", name, err)
			}
		})
	}
	if err := store.RemoveAll("."); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("RemoveAll(.) error = %v, want ErrInvalidPath", err)
	}
}

func TestReadOnlyStorageRejectsWrites(t *testing.T) {
	base, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	if err := base.WriteFile("a.txt", []byte("content")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	readonly, err := NewReadOnly(base)
	if err != nil {
		t.Fatalf("NewReadOnly() error = %v", err)
	}

	if got, err := readonly.ReadFile("a.txt"); err != nil || string(got) != "content" {
		t.Fatalf("ReadFile() = %q, %v", got, err)
	}
	if err := readonly.WriteFile("b.txt", []byte("blocked")); err == nil {
		t.Fatal("WriteFile() on read-only storage error is nil")
	}
	if err := readonly.Remove("a.txt"); err == nil {
		t.Fatal("Remove() on read-only storage error is nil")
	}
	if err := readonly.Copy("a.txt", "b.txt"); err == nil {
		t.Fatal("Copy() on read-only storage error is nil")
	}
}

func TestCopyMemoryDirectory(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	if err := store.WriteFile("src/a.txt", []byte("a")); err != nil {
		t.Fatalf("WriteFile(a) error = %v", err)
	}
	if err := store.WriteFile("src/nested/b.txt", []byte("b")); err != nil {
		t.Fatalf("WriteFile(b) error = %v", err)
	}

	if err := store.Copy("src", "dst"); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if got, err := store.ReadFile("dst/nested/b.txt"); err != nil || string(got) != "b" {
		t.Fatalf("ReadFile(dst/nested/b.txt) = %q, %v", got, err)
	}
	if err := store.Copy("src", "dst"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("Copy() existing error = %v, want ErrAlreadyExists", err)
	}
	if err := store.WriteFile("src/nested/b.txt", []byte("updated")); err != nil {
		t.Fatalf("WriteFile(update) error = %v", err)
	}
	if err := store.Copy("src", "dst", WithOverwrite(true)); err != nil {
		t.Fatalf("Copy(overwrite) error = %v", err)
	}
	if got, err := store.ReadFile("dst/nested/b.txt"); err != nil || string(got) != "updated" {
		t.Fatalf("ReadFile(overwritten) = %q, %v", got, err)
	}
}

func TestCopyOSDirectory(t *testing.T) {
	store, err := NewOS(t.TempDir())
	if err != nil {
		t.Fatalf("NewOS() error = %v", err)
	}
	if err := store.WriteFile("src/a.txt", []byte("a")); err != nil {
		t.Fatalf("WriteFile(a) error = %v", err)
	}
	if err := store.WriteFile("src/nested/b.txt", []byte("b")); err != nil {
		t.Fatalf("WriteFile(b) error = %v", err)
	}

	if err := store.Copy("src", "dst"); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if got, err := store.ReadFile("dst/nested/b.txt"); err != nil || string(got) != "b" {
		t.Fatalf("ReadFile(dst/nested/b.txt) = %q, %v", got, err)
	}
	if err := store.Copy("src", "dst"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("Copy() existing error = %v, want ErrAlreadyExists", err)
	}
	if err := store.WriteFile("src/nested/b.txt", []byte("updated")); err != nil {
		t.Fatalf("WriteFile(update) error = %v", err)
	}
	if err := store.Copy("src", "dst", WithOverwrite(true)); err != nil {
		t.Fatalf("Copy(overwrite) error = %v", err)
	}
	if got, err := store.ReadFile("dst/nested/b.txt"); err != nil || string(got) != "updated" {
		t.Fatalf("ReadFile(overwritten) = %q, %v", got, err)
	}
}

func TestDetectMIME(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	if err := store.WriteFile("image.bin", png); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	info, err := store.DetectMIME("image.bin")
	if err != nil {
		t.Fatalf("DetectMIME() error = %v", err)
	}
	if info.MIME != "image/png" || info.Extension != ".png" {
		t.Fatalf("DetectMIME() = %#v", info)
	}

	textInfo := DetectBytes([]byte("hello"))
	if !strings.HasPrefix(textInfo.MIME, "text/plain") {
		t.Fatalf("DetectBytes() = %#v", textInfo)
	}
}

func TestWatchOSStorage(t *testing.T) {
	store, err := NewOS(t.TempDir(), WithWatchDebounce(10*time.Millisecond))
	if err != nil {
		t.Fatalf("NewOS() error = %v", err)
	}
	if err := store.MkdirAll("watch"); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	watcher, err := store.Watch("watch", WithRecursiveWatch(true), WithEventDebounce(10*time.Millisecond))
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}
	defer watcher.Close()

	if err := store.WriteFile("watch/file.txt", []byte("hello")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	createdOrWritten := waitForEvent(t, watcher, func(event Event) bool {
		return event.Path == "watch/file.txt" && (event.Op == WatchOpCreate || event.Op == WatchOpWrite)
	})
	if createdOrWritten.Path != "watch/file.txt" {
		t.Fatalf("unexpected event = %#v", createdOrWritten)
	}

	if err := store.Remove("watch/file.txt"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	removed := waitForEvent(t, watcher, func(event Event) bool {
		return event.Path == "watch/file.txt" && (event.Op == WatchOpRemove || event.Op == WatchOpRename)
	})
	if removed.Path != "watch/file.txt" {
		t.Fatalf("unexpected remove event = %#v", removed)
	}
}

func TestWatchMemoryStorageUnsupported(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	_, err = store.Watch(".")
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Watch() error = %v, want ErrUnsupported", err)
	}
}

func TestStorageConcurrentAccess(t *testing.T) {
	store, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := filepath.ToSlash(filepath.Join("parallel", "file-"+string(rune('a'+i))+".txt"))
			if err := store.WriteFile(name, []byte(name)); err != nil {
				t.Errorf("WriteFile(%s) error = %v", name, err)
				return
			}
			got, err := store.ReadFile(name)
			if err != nil {
				t.Errorf("ReadFile(%s) error = %v", name, err)
				return
			}
			if string(got) != name {
				t.Errorf("ReadFile(%s) = %q", name, got)
			}
		}()
	}
	wg.Wait()
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func waitForEvent(t *testing.T, watcher *Watcher, match func(Event) bool) Event {
	t.Helper()
	timeout := time.After(3 * time.Second)
	for {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				t.Fatal("watcher events channel closed")
			}
			if match(event) {
				return event
			}
		case err := <-watcher.Errors():
			if err != nil {
				t.Fatalf("watcher error = %v", err)
			}
		case <-timeout:
			t.Fatal("timed out waiting for watcher event")
		}
	}
}
