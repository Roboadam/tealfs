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
	"tealfs/pkg/model"
	"time"

	"golang.org/x/net/webdav"
)

type FileSystem struct {
	FilesByPath    map[string]File
	FilesByBlockId map[model.BlockId]File
	BlockRequest   chan model.BlockId
	BlockResponse  chan blockResponse
	addFile        chan addFileReq
	openFileReq    chan openFileReq
}

func (f *FileSystem) run() {
	for {
		select {
		case af := <-f.addFile:
			f.FilesByPath[af.name] = af.file
			if af.file.BlockId != "" {
				f.FilesByBlockId[af.file.BlockId] = af.file
			}
		case of := <-f.openFileReq:
			f, exists := f.FilesByPath[of.name]
			if exists {
				of.resp <- openFileResp{file: f}
			} else {
				of.resp <- openFileResp{err: errors.New("file not found")}
			}
		}
	}
}

type blockResponse struct {
	data []byte
	id   model.BlockId
}

type addFileReq struct {
	name string
	file File
}

type openFileReq struct {
	name string
	resp chan openFileResp
}

type openFileResp struct {
	file File
	err  error
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// Todo handle the context

	_, exists := f.FilesByPath[name]
	if exists {
		return errors.New("path exists path")
	}

	dirName, fileName := dirAndFileName(name)
	_, exists = f.FilesByPath[dirName]
	if !exists {
		return errors.New("invalid path")
	}

	dir := File{
		NameValue:    fileName,
		IsDirValue:   true,
		RO:           false,
		RW:           false,
		WO:           false,
		Append:       false,
		Create:       false,
		FailIfExists: exists,
		Truncate:     false,
		SizeValue:    0,
		ModeValue:    0,
		Modtime:      time.Now(),
		SysValue:     nil,
		Position:     0,
		Data:         []byte{},
		IsOpen:       false,
		BlockId:      "",
	}

	f.addFile <- addFileReq{
		name: name,
		file: dir,
	}

	return nil
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return f.openFile(name, flag, perm)
}

func (f *FileSystem) openFile(name string, flag int, perm os.FileMode) (*File, error) {
	ro := os.O_RDONLY&flag != 0
	rw := os.O_RDWR&flag != 0
	wo := os.O_WRONLY&flag != 0
	append := os.O_APPEND&flag != 0
	create := os.O_CREATE&flag != 0
	failIfExists := os.O_EXCL&flag != 0
	truncate := os.O_TRUNC&flag != 0
	isDir := perm.IsDir()

	// only one of ro, rw, wo allowed
	if (ro && rw) || (ro && wo) || (rw && wo) || !(ro || rw || wo) {
		return nil, errors.New("invalid flag")
	}

	if ro && (append || create || failIfExists || truncate) {
		return nil, errors.New("invalid flag")
	}

	if !create && failIfExists {
		return nil, errors.New("invalid flag")
	}

	// opening the root directory
	dirName, fileName := dirAndFileName(name)
	if fileName == "" && dirName == "" {
		if isDir {
			resp := make(chan openFileResp)
			f.openFileReq <- openFileReq{name: "", resp: resp}
			of := <-resp
			return &of.file, of.err
		} else {
			return nil, errors.New("not a directory")
		}
	}

	// make sure parent directory is valid
	resp := make(chan openFileResp)
	f.openFileReq <- openFileReq{name: dirName, resp: resp}
	of := <- resp
	if of.err != nil {
		return nil, of.err
	}

	
	for _, dirName := range dirNames {
		if !current.IsDirValue {
			return nil, errors.New("invalid path")
		}

		dir, exists := current.Chidren[dirName]
		if !exists {
			return nil, errors.New("invalid path")
		}
		current = dir
	}

	file, exists := current.Chidren[*fileName]

	if exists && failIfExists {
		return nil, errors.New("invalid path")
	}

	current.RO = ro
	current.RW = rw
	current.WO = wo
	current.Append = append
	current.Create = create
	current.FailIfExists = failIfExists
	current.Truncate = truncate
	if append && !ro {
		current.Position = current.SizeValue
	} else {
		current.Position = 0
	}

	return &current, nil
}

func last(i int, arry []string) bool {
	return i == len(arry)-1
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	pathsArry := paths(name)
	parentName := strings.Join(pathsArry[:len(pathsArry)-1], "/")

	file, err := f.openFile(parentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	delete(file.Chidren, pathsArry[len(pathsArry)-1])

	return nil
}

func (f *FileSystem) Rename(ctx context.Context, oldName string, newName string) error {
	oldPathsArry := paths(oldName)
	oldParentName := strings.Join(oldPathsArry[:len(oldPathsArry)-1], "/")
	oldParent, err := f.openFile(oldParentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	oldSimpleName := oldPathsArry[len(oldPathsArry)-1]
	file, exists := oldParent.Chidren[oldSimpleName]
	if !exists {
		return errors.New("file does not exist")
	}

	newPathsArry := paths(newName)
	newParentName := strings.Join(newPathsArry[:len(newPathsArry)-1], "/")

	newParent, err := f.openFile(newParentName, os.O_RDWR, os.ModeDir)
	if err != nil {
		return err
	}

	file.NameValue = newPathsArry[len(newPathsArry)-1]
	delete(oldParent.Chidren, oldSimpleName)
	newParent.Chidren[file.NameValue] = file

	return nil
}

func dirAndFileName(name string) (string, string) {
	raw := strings.Split(name, "/")
	result := make([]string, 0)
	for _, value := range raw {
		if value != "" {
			result = append(result, value)
		}
	}
	last := len(result) - 1
	if last < 0 {
		return "", ""
	}
	return strings.Join(result[:last], "/"), result[last]
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Todo. don't know what right mode is here
		return f.openFile(name, os.O_RDONLY, os.ModeExclusive)
	}
}

type File struct {
	NameValue    string
	IsDirValue   bool
	RO           bool
	RW           bool
	WO           bool
	Append       bool
	Create       bool
	FailIfExists bool
	Truncate     bool
	SizeValue    int64
	ModeValue    fs.FileMode
	Modtime      time.Time
	SysValue     any
	Position     int64
	Data         []byte
	IsOpen       bool
	BlockId      model.BlockId
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
	return f, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (f *File) Name() string {
	return f.NameValue
}

func (f *File) Size() int64 {
	return f.SizeValue
}

func (f *File) Mode() fs.FileMode {
	return f.ModeValue
}

func (f *File) ModTime() time.Time {
	return f.Modtime
}

func (f *File) IsDir() bool {
	return f.IsDirValue
}

func (f *File) Sys() any {
	return f.SysValue
}
