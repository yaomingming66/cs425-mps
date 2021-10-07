# CS 425

- [MP0](./docs/mp0/README.md)
- [MP1](./docs/mp1/README.md)
- [MP2](./docs/mp2/README.md)
- [MP3](./docs/mp3/README.md)

## Development Environment configuration

### Golang

[Install](https://golang.org/doc/install)

```bash
# version >= 1.13
go version
go get golang.org/x/tools/cmd/goimports
go get -u github.com/rakyll/gotest
```

### VSCode

```bash
# extension id golang.go
code --install-extension golang.go
```

### Commonly used quick commands

fmt code

```bash
bash ./script/unix/fmt.bash
```

```linux
./script/linux/fmt.bash
```

build mp0 static binary to ./bin

```bash
bash ./script/unix/mp0/build.bash
```

```linux
./script/linux/mp0/build.sh
```

```bash
# just stderr
mp0-s 8080 1> /dev/null
# just stdout
mp0-s 8080 2> /dev/null
```

Find this list of possible platforms

```bash
go tool dist list
```

test Spec

```bash
gotest -mod vendor -v ./<>
```

test All

```bash
gotest -mod vendor -v ./...
```

trace log

```bash
LOG=trace
```

json log

```bash
LOG=json
```
