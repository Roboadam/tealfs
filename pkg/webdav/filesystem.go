// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package webdav

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"time"

	"golang.org/x/net/webdav"
)

type FileSystem struct {
	Root File
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// Todo handle the context
	names := paths(name)
	current := f.Root
	for _, dir := range names[:len(names)-1] {
		if hasDirWithName(&current, dir) {
			current = current.Chidren[dir]
		} else {
			return errors.New("invalid path")
		}
	}
	if hasChildWithName(&current, names[len(names)-1]) {
		return errors.New("invalid path")
	}
	current.Chidren[names[len(names)-1]] = File{
		Name:    names[len(names)-1],
		IsDir:   true,
		Chidren: map[string]File{},
	}
	return nil
}

func hasDirWithName(file *File, dirName string) bool {
	child, exists := file.Chidren[dirName]
	return exists && child.IsDir
}

func hasChildWithName(file *File, dirName string) bool {
	_, exists := file.Chidren[dirName]
	return exists
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	ro := os.O_RDONLY&flag != 0
	rw := os.O_RDWR&flag != 0
	wo := os.O_WRONLY&flag != 0
	append := os.O_APPEND&flag != 0
	create := os.O_CREATE&flag != 0
	failIfExists := os.O_EXCL&flag != 0
	truncate := os.O_TRUNC&flag != 0

	pathNames := paths(name)
	current := f.Root
	for i, pathName := range pathNames {
		if !current.IsDir {
			return nil, errors.New("invalid path")
		}

		if last(i, pathNames) {
			_, exists := current.Chidren[pathName]
			if create && failIfExists && exists {
				return nil, errors.New("file exists")
			}
		}

		file, exists := current.Chidren[pathName]
		if !exists {
			return nil, errors.New("file doesn't exist")
		}
		current = file
	}
	return &current, nil
}

func last(i int, arry []string) bool {
	return i == len(arry)-1
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	panic("not implemented") // TODO: Implement
}

func (f *FileSystem) Rename(ctx context.Context, oldName string, newName string) error {
	panic("not implemented") // TODO: Implement
}

func paths(name string) []string {
	raw := strings.Split(name, "/")
	result := make([]string, 0)
	for _, value := range raw {
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		info := FileInfo{
			name:    name,
			size:    0,
			mode:    0,
			modtime: time.Time{},
			isdir:   false,
			sys:     nil,
		}
		return &info, nil
	}
}

type File struct {
	Name    string
	IsDir   bool
	Chidren map[string]File
}

func (f *File) Close() error {
	return nil
}

func (f *File) Read(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Stat() (fs.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Write(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modtime time.Time
	isdir   bool
	sys     any
}

func (f *FileInfo) Name() string {
	return f.name
}

func (f *FileInfo) Size() int64 {
	return f.size
}

func (f *FileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.modtime
}

func (f *FileInfo) IsDir() bool {
	return f.isdir
}

func (f *FileInfo) Sys() any {
	return f.sys
}
