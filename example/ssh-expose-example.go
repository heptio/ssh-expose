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

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	sshexpose "github.com/heptio/ssh-expose"
)

var serveSSH = flag.String("serve-ssh", "", "Specify an address to serve SSH on instead of std*")

func main() {
	flag.Parse()
	if len(*serveSSH) > 0 {
		opt := sshexpose.SSHOptions{
			ICLI:           runICLI,
			Addr:           *serveSSH,
			PrivateHostKey: "./example_host_key_rsa",
		}
		err := sshexpose.ServeSSH(opt)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		sshexpose.ServeLocal(runICLI)
	}
}

func runICLI(t sshexpose.Term) {
	outWriter := t.StdOut()
	errWriter := t.StdErr()

	var text string
	var err error
	for {
		fmt.Fprint(errWriter, "> ")
		text, err = t.ReadLine()
		if err != nil {
			break
		}

		if strings.ToLower(text) == "quit" {
			return
		}

		fmt.Fprintf(outWriter, "Server: %s\n", text)
	}
	if err != nil && err != io.EOF {
		log.Printf("Error reading from stream: %v", err)
	}
}
