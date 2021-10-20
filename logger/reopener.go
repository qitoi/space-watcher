/*
 *  Copyright 2021 qitoi
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package logger

import (
	"os"
	"sync"
)

type WriteSyncer interface {
	Write(p []byte) (int, error)
	Sync() error
}

type WriteSyncReopener interface {
	Write(p []byte) (int, error)
	Sync() error
	Reopen() error
}

func Wrap(ws WriteSyncer) WriteSyncReopener {
	return &nilReopener{
		WriteSyncer: ws,
	}
}

type nilReopener struct {
	WriteSyncer
}

func (r *nilReopener) Reopen() error {
	return nil
}

type fileReopener struct {
	mu   sync.Mutex
	f    *os.File
	name string
	flag int
	perm os.FileMode
}

func OpenFile(name string, flag int, perm os.FileMode) (WriteSyncReopener, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &fileReopener{
		f:    f,
		name: name,
		flag: flag,
		perm: perm,
	}, nil
}

func (fr *fileReopener) Write(p []byte) (int, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	return fr.f.Write(p)
}

func (fr *fileReopener) Sync() error {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	return fr.f.Sync()
}

func (fr *fileReopener) Reopen() error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	fr.f.Close()

	f, err := os.OpenFile(fr.name, fr.flag, fr.perm)
	if err != nil {
		return err
	}

	fr.f = f
	return nil
}
