#!/bin/bash

while read line
do
        echo "$line"
        cd ../$line
        go test ./... -cover
        cd -
done < test-suite
