#!/bin/sh

clear
while [ 1 ] 
do
    #inotifywait -q -e modify -e create -e delete . > /dev/null
    inotifywait -q -e modify *.go > /dev/null
    clear
    #go build && scp lightserver pi:
    #go build
    go test
done
