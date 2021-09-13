#!/usr/bin/env bash

set -o pipefail
set -o errexit
set -o xtrace

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

sleep 10 && python3  ./script/unix/mp0/generator.py 10000 100000 | ./bin/mp0-c A 127.1 8080 &
sleep 10 && python3  ./script/unix/mp0/generator.py 10000 100000 | ./bin/mp0-c B 127.1 8080 &
sleep 10 && python3  ./script/unix/mp0/generator.py 10000 100000 | ./bin/mp0-c C 127.1 8080 &
sleep 10 && python3  ./script/unix/mp0/generator.py 10000 100000 | ./bin/mp0-c D 127.1 8080 &

./bin/mp0-s 8080 2> /tmp/a.log