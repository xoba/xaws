package xaws

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"
)

type fakeFile struct {
	size    int64
	statErr error
}

func (f fakeFile) Read(p []byte) (int, error) { return 0, io.EOF }

func (f fakeFile) Stat() (fs.FileInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	return fakeFileInfo{size: f.size}, nil
}

type fakeFileInfo struct{ size int64 }

func (fakeFileInfo) Name() string       { return "" }
func (fi fakeFileInfo) Size() int64     { return fi.size }
func (fakeFileInfo) Mode() fs.FileMode  { return 0 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() any           { return nil }

func TestUploadMultipartStatError(t *testing.T) {
	f := fakeFile{statErr: errors.New("boom")}
	_, err := UploadMultipart(context.Background(), nil, f, "b", "k")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("want boom error, got %v", err)
	}
}

func TestUploadMultipartFileTooBig(t *testing.T) {
	f := fakeFile{size: MaxTotalSize + 1}
	_, err := UploadMultipart(context.Background(), nil, f, "b", "k")
	if err == nil || !strings.Contains(err.Error(), "file too big") {
		t.Fatalf("expected file too big error, got %v", err)
	}
}
