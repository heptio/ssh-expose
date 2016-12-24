// Copyright 2016 Heptio Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sshexpose

import (
	"bufio"
	"io"
	"os"
)

type localTerm struct {
	scanner *bufio.Scanner
}

func (t *localTerm) ReadLine() (line string, err error) {
	more := t.scanner.Scan()
	if !more {
		if t.scanner.Err() != nil {
			return "", t.scanner.Err()
		}
		return "", io.EOF
	}
	return t.scanner.Text(), nil
}

func (t *localTerm) StdOut() io.Writer {
	return os.Stdout
}

func (t *localTerm) StdErr() io.Writer {
	return os.Stdout
}

func (t *localTerm) Close() {}

// ServeLocal will serve an interactive CLI locally.  This takes care of
// providing a line based input method.
func ServeLocal(icli ServeICLI) {
	t := localTerm{
		scanner: bufio.NewScanner(os.Stdin),
	}
	icli(&t)
}
