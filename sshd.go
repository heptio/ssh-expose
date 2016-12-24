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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type sshTerm struct {
	t *terminal.Terminal
	c ssh.Channel
}

func (t *sshTerm) ReadLine() (line string, err error) {
	return t.t.ReadLine()
}

func (t *sshTerm) StdOut() io.Writer {
	return t.t
}

func (t *sshTerm) StdErr() io.Writer {
	return t.t
}

func (t *sshTerm) Close() {
	t.c.Close()
}

// SSHOptions defines the parameters for serving a REPL over SSH
type SSHOptions struct {
	// ICLI defines the interactive CLI you want to serve
	ICLI ServeICLI

	// Addr is an IPv4 IP:port pair that you wish to serve on
	Addr string

	// PrivateHostKey is the host key to present when serving SSH.
	//
	// This can be generated with `ssh-keygen -t rsa -f my_host_key_rsa`
	PrivateHostKey string
}

type sshServer struct {
	icli ServeICLI
}

// ServeSSH will run an SSH server that accepts connections.
//
// Each incoming SSH session will result in a call to options.REPL.
func ServeSSH(options SSHOptions) error {
	srv := &sshServer{
		icli: options.ICLI,
	}

	log.Printf("Serving SSH on %v\n", options.Addr)
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	privateBytes, err := ioutil.ReadFile(options.PrivateHostKey)
	if err != nil {
		return errors.Wrapf(err, "Failed to load private key (%s)", options.PrivateHostKey)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse private key (%s)", options.PrivateHostKey)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", options.Addr)
	if err != nil {
		return errors.Wrapf(err, "Failed to listen on %s", options.Addr)
	}

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go srv.handleChannels(chans)
	}
}

func (srv *sshServer) handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go srv.handleChannel(newChannel)
	}
}

type ptyReqMsg struct {
	Term        string
	CharWidth   uint32
	CharHeight  uint32
	PixelWidth  uint32
	PixelHeight uint32
	Modes       string
}

type windowChangeMsg struct {
	CharWidth   uint32
	CharHeight  uint32
	PixelWidth  uint32
	PixelHeight uint32
}

func (srv *sshServer) handleChannel(newChannel ssh.NewChannel) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	defer connection.Close()

	t := &sshTerm{
		t: terminal.NewTerminal(connection, ""),
		c: connection,
	}

	// Sessions have out-of-band requests such as "shell",
	// "pty-req" and "env".  Here we handle only the
	// "shell" request.
	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "shell":
				req.Reply(true, nil)
			case "pty-req":
				var ptyReq ptyReqMsg
				err := ssh.Unmarshal(req.Payload, &ptyReq)
				if err != nil {
					log.Printf("Error parsing 'pty-req' payload")
					req.Reply(false, nil)
				}
				t.t.SetSize(int(ptyReq.CharWidth), int(ptyReq.CharHeight))
				req.Reply(true, nil)
			case "window-change":
				var windowChange windowChangeMsg
				err := ssh.Unmarshal(req.Payload, &windowChange)
				if err != nil {
					log.Printf("Error parsing 'window-change' payload")
				}
				t.t.SetSize(int(windowChange.CharWidth), int(windowChange.CharHeight))
			}
		}
	}(requests)

	srv.icli(t)
}
