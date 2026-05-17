package storage

import (
	"path"
	"path/filepath"
	"strings"
)

func cleanStoragePath(name string) (string, error) {
	if name == "" {
		return "", invalidPath(name)
	}
	if filepath.IsAbs(name) || filepath.VolumeName(name) != "" || hasWindowsVolume(name) {
		return "", invalidPath(name)
	}

	slashName := strings.ReplaceAll(name, "\\", "/")
	if path.IsAbs(slashName) {
		return "", invalidPath(name)
	}

	cleaned := path.Clean(slashName)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", invalidPath(name)
	}
	return filepath.FromSlash(cleaned), nil
}

func hasWindowsVolume(name string) bool {
	if len(name) < 2 || name[1] != ':' {
		return false
	}
	first := name[0]
	return (first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z')
}

func rejectRootMutation(name string) error {
	if filepath.Clean(name) == "." {
		return invalidPath(name)
	}
	return nil
}

func publicPath(name string) string {
	return filepath.ToSlash(filepath.Clean(name))
}
