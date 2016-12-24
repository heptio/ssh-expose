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

import "io"

// Term is a simple terminal abstraction.
//
// This is what programs should use for all input/output with a user.  It will
// work equally well over SSH or the local terminal.
type Term interface {
	ReadLine() (line string, err error)
	StdOut() io.Writer
	StdErr() io.Writer
	Close()
}

// ServeICLI should implement interactive CLI
type ServeICLI func(t Term)
