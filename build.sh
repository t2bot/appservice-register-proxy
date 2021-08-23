#!/bin/bash

GOOS=linux GOARCH=amd64 GOBIN=$PWD/bin go build -o $PWD/bin/linux-amd64-as_proxy -a -v ./main.go
