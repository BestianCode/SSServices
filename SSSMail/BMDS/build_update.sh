#!/bin/sh

export GOPATH=$HOME/go
export GOROOT=/opt/go
export PATH=$PATH:~/bin:/opt/go/bin:$GOPATH/bin

git pull
go build

