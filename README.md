# CS 425

## Development Environment configuration

### Golang

[Install](https://golang.org/doc/install)

```bash
# version >= 1.13
go version
```

### VSCode

```bash
# extension id golang.go
code --install-extension golang.go
```

## MP0

### Commonly used quick commands

```bash
# build mp0 static binary to ./bin
bash ./script/mp0/build_mp0.bash
```

```bash
# just stderr
mp0-s 8080 1> /dev/null
# just stdout
mp0-s 8080 2> /dev/null
```
