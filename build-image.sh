!/usr/bin/env bash

set -e -x

SYSLGEN_BRANCH=master

function cleanup() {
  rm -rf sysl syslgen-examples
}
trap cleanup EXIT

go get -t -v github.com/anz-bank/sysl/sysl2/sysl
python setup.py sdist bdist_wheel
GOOS=darwin GOARCH=amd64 go build -o gosysl/gosysl-darwin github.com/anz-bank/sysl/sysl2/sysl
GOOS=linux GOARCH=amd64 go build -o gosysl/gosysl-linux github.com/anz-bank/sysl/sysl2/sysl

git clone https://github.com/anz-bank/syslgen-examples.git -b $SYSLGEN_BRANCH

docker build -t sysl2 -f Dockerfile-gosysl .
