## ssh-expose

A simple package that makes it easy to write a interactive command line program that can be used locally or served over SSH.

## Example
See the `example` directory for an .... example.

```
cd example
# Local mode
go run *.go

# SSH Mode
ssh-keygen -t rsa -f example_host_key_rsa
go run *.go -serve-ssh=localhost:2222
# Run 'ssh localhost -p 2222' on another terminal
```

## TODO

* [ ] Support authentication to sshd
* [ ] Rate limiting and limite the num of connections
* [ ] readline type stuff with tab completion, etc
