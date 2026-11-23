package confab

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// File is an uploaded file once stored.
type File struct {
	Key, Name, ContentType string
	Size                   int64
}

// Files keeps uploads in a directory, each under a random key.
type Files struct{ dir string }

func NewFiles(dir string) (*Files, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Files{dir: dir}, nil
}

var keyPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// Put stores an upload and returns it with its new key.
func (f *Files) Put(h *multipart.FileHeader) (File, error) {
	src, err := h.Open()
	if err != nil {
		return File{}, err
	}
	defer src.Close()
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	key := hex.EncodeToString(b)
	dst, err := os.OpenFile(filepath.Join(f.dir, key), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return File{}, err
	}
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ct := http.DetectContentType(head[:n])
	size, err := io.Copy(dst, io.MultiReader(bytes.NewReader(head[:n]), src))
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(dst.Name())
		return File{}, err
	}
	return File{Key: key, Name: filepath.Base(h.Filename), ContentType: ct, Size: size}, nil
}

// Open opens a stored file; ErrNotFound for a key that isn't one.
func (f *Files) Open(key string) (*os.File, error) {
	if !keyPattern.MatchString(key) {
		return nil, ErrNotFound
	}
	file, err := os.Open(filepath.Join(f.dir, key))
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return file, err
}

// Delete removes stored files; a missing one is no error.
func (f *Files) Delete(keys ...string) error {
	for _, k := range keys {
		if !keyPattern.MatchString(k) {
			return fmt.Errorf("files: bad key %q", k)
		}
		if err := os.Remove(filepath.Join(f.dir, k)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
