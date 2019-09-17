#!/bin/bash

set -e

BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%SZ'`
GO_VERSION=`go version`
CMD="$1"
VERSION="$2"
COMMIT="$3"
GOOS="$4"
GOARCH="$5"
BUILD_OS="$6"
OUT="$7"

usage(){
    echo "Usage: $0 <build|install> <MAJOR.MINOR.PATCH> <commit> <go-os> <go-arch> <build-os> <out-file(only for build)>"
    echo "eg: $0 build 1.0.0 cfe447 linux amd64 darwin out/gosysl-darwin"
}

if [[ -z ${CMD} ]]; then
    echo "Command is not specified"
    usage
    exit 1
elif [[ ${CMD} != "build" && ${CMD} != "install"  ]]; then
    echo "Invalid command"
    usage
    exit 1
elif [[ -z ${VERSION} ]]; then
    echo "Version is not specified"
    usage
    exit 1
elif ! [[ ${VERSION} =~ ^[[:digit:]+\\.[:digit:]+\\.[:digit:]]+$ ]]; then
    echo "Version is invalid. Binary version will be empty"
    VERSION=""
elif [[ -z ${COMMIT} ]]; then
    echo "Commit SHA is not specified"
    usage
    exit 1
elif [[ -z ${GOOS} ]]; then
    echo "Go OS is not specified"
    usage
    exit 1
elif [[ -z ${GOARCH} ]]; then
    echo "Go Arch is not specified"
    usage
    exit 1
elif [[ -z ${BUILD_OS} ]]; then
    echo "Build OS is not specified"
    usage
    exit 1
elif [[ ${CMD} == "build" && -z ${OUT} ]]; then
    echo "Output is not specified"
    usage
    exit 1
fi

FLAGS="\
    -X \"main.Version=${VERSION}\" \
    -X \"main.GitCommit=${COMMIT}\" \
    -X \"main.BuildDate=${BUILD_DATE}\" \
    -X \"main.GoOS=${GOOS}\" \
    -X \"main.GoArch=${GOARCH}\" \
    -X \"main.GoVersion=${GO_VERSION}\" \
    -X \"main.BuildOS=${BUILD_OS}\"" \

if [[ ${CMD} = "build" ]]; then
    GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${OUT} -ldflags "${FLAGS}" -v github.com/anz-bank/sysl/sysl2/sysl
elif [[ ${CMD} = "install" ]]; then
    go install -ldflags "${FLAGS}" -v ./sysl2/sysl
fi
