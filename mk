#!/usr/bin/env bash

function deps {
echo "Fetching dependencies to $GOPATH..."
printf "   (00/03)\r"
  go get -u github.com/stretchr/testify
printf "   (01/03)\r"
  go get -u github.com/ivotron/peg
printf "   (02/03)\r"
  go get -u github.com/mattn/go-sqlite3
printf "## (03/03)\r"
printf "\n"
}

function build {
  go build ./...
}

function install {
  go install ./...
}

function test {
  go test ./...
}

$1
