#!/bin/bash

set -e

BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%SZ'`
GO_VERSION=`go version`
CMD="$1"
VERSION="$2"
COMMIT="$3"
BUILD_OS="$4"
OUT="$5"
GOOS="$6"
GOARCH="$7"

usage(){
    echo "Usage: $0 <build|install> <MAJOR.MINOR.PATCH> <commit> <build-os> <out-file(only for build)> [go-os(only for build)] [go-arch(only for build)]"
    echo "eg: $0 build 1.0.0 cfe447 darwin out/gosysl-darwin linux amd64"
}

if [[ -z ${CMD} ]]; then
    echo "Command is not specified"
    usage
    exit 1
fi
if [[ ${CMD} != "build" && ${CMD} != "install"  ]]; then
  echo "Invalid command"
  usage
  exit 1
fi
if [[ -z ${VERSION} ]]; then
  echo "Version is not specified"
  usage
  exit 1
fi
if ! [[ ${VERSION} =~ ^([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  echo "Version is invalid. Binary version will be empty"
  VERSION=""
fi
if [[ -z ${COMMIT} ]]; then
  echo "Commit SHA is not specified"
  usage
  exit 1
fi
if [[ -z ${BUILD_OS} ]]; then
  echo "Build OS is not specified"
  usage
  exit 1
fi
if [[ ${CMD} == "build" && -z ${OUT} ]]; then
  echo "Output is not specified"
  usage
  exit 1
fi
if [[ ${CMD} == "build" && (-z ${GOOS} || -z ${GOARCH}) ]]; then
  echo "GO OS and/or GO Arch is not specified. Default go build command will be used"
fi

FLAGS="\
    -X \"main.Version=${VERSION}\" \
    -X \"main.GitCommit=${COMMIT}\" \
    -X \"main.BuildDate=${BUILD_DATE}\" \
    -X \"main.GoVersion=${GO_VERSION}\" \
    -X \"main.BuildOS=${BUILD_OS}\"" \

if [[ ${CMD} = "build" ]]; then
    if [[ -n ${GOOS} && -n ${GOARCH} ]]; then
        GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${OUT} -ldflags "${FLAGS}" -v github.com/anz-bank/sysl/sysl2/sysl
    else
        go build -o ${OUT} -ldflags "${FLAGS}" -v github.com/anz-bank/sysl/sysl2/sysl
    fi
elif [[ ${CMD} = "install" ]]; then
    go install -ldflags "${FLAGS}" -v ./sysl2/sysl
fi
