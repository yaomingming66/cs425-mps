#!/usr/bin/env bash

set -o pipefail
set -o errexit
set -o xtrace

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

sleep 10 && python3 -u ./script/unix/mp0/generator.py 0.5 50 | ./bin/mp0-c A 127.1 8080 &
sleep 10 && python3 -u ./script/unix/mp0/generator.py 0.5 50 | ./bin/mp0-c B 127.1 8080 &
sleep 10 && python3 -u ./script/unix/mp0/generator.py 0.5 50 | ./bin/mp0-c C 127.1 8080 &

./bin/mp0-s 8080 2> ./visualization/mp0/3_node_0_5_hz_100_s.log