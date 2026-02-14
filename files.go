package doze

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Files manipulation utilities.

// Copy a file from `src` to `dst`.
// The `dst` directory tree must exist before attempting to copy the file.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return nil
}

// Create the base directory tree of `dst`.
// Ex:
// - dst = `/foo/bar/boz.txt` will create `/foo/bar`.
// - dst = `boz.txt` will not create anything.
func CreateBaseDir(dst string) error {
	baseDir := filepath.Dir(dst)
	if baseDir != "." {
		if err := os.MkdirAll(baseDir, 0o775); err != nil {
			return fmt.Errorf("failed to create base directory tree at %s: %s", baseDir, err)
		}
	}

	return nil
}
