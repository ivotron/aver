#!/usr/bin/env bash

function deps {
echo "Fetching dependencies to $GOPATH..."
printf "   (00/02)\r"
  go get -u github.com/stretchr/testify
printf "   (01/02)\r"
  go get -u github.com/jarcoal/httpmock
printf "## (02/02)\r"
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
