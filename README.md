# CS 425

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

## MP0

```bash
# Client Test
./bin/mp0-s 8080 or socat TCP-LISTEN:8080 -
./bin/mp0-c A 127.0.0.1 8080
python3 ./script/mp0/generator.py 1 100 | ./bin/mp0-c A 127.0.0.1 8080
```

## MP1

```bash
# Quick Build
bash ./script/unix/mp1/quick_build.bash
# Release Build
bash ./script/unix/mp1/build.bash
# Usage
./bin/mp1 a 8080 ./lib/mp1/config/config_a.txt
./bin/mp1 b 8081 ./lib/mp1/config/config_b.txt
./bin/mp1 c 8082 ./lib/mp1/config/config_c.txt

python3 -u gentx.py 0.5 | ./bin/mp1 a 8080 ./lib/mp1/config/config_a.txt
python3 -u gentx.py 0.5 | ./bin/mp1 b 8081 ./lib/mp1/config/config_b.txt
python3 -u gentx.py 0.5 | ./bin/mp1 c 8082 ./lib/mp1/config/config_c.txt
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
