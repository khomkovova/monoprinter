#!/bin/bash

go get -d ./...
go build
./MonoPrinter & echo "Program runing with process id = "$!
PID=$!
dlv attach --headless --listen=:2345 $PID `pwd` && kill $PID


