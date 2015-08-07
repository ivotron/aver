#!/usr/bin/env bash

function deps {
echo "Fetching dependencies to $GOPATH..."
printf "   (00/05)\r"
  go get -u github.com/stretchr/testify
printf "   (01/05)\r"
  go get -u github.com/ivotron/peg
printf "   (02/05)\r"
  go get -u github.com/mattn/go-sqlite3
printf "   (03/05)\r"
  go get -u github.com/spf13/cobra
printf "   (04/05)\r"
  go get -u github.com/bitly/go-simplejson
printf "## (05/05)\r"
printf "\n"
}

$1
