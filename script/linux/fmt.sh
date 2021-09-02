#!/bin/bash

gofmt -s -w -l -d ./cli ./cmd ./lib
goimports -w -l -d ./cli ./cmd ./lib
echo "Formated"