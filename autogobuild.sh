#!/bin/sh

clear
while [ 1 ] 
do
    clear
    #go build && scp lightserver pi:
    go build && echo "go build OK"
    #go test && echo "go test OK"
    #inotifywait -q -e modify -e create -e delete . > /dev/null
    inotifywait -q -e modify *.go > /dev/null
done
