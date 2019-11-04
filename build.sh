#!/bin/bash
set -eo pipefail
shopt -s nullglob

cd src/server
glide update
cd ../..

if [ ! -z "$1" ]; then
    tag=$1
    sudo docker build -t jinnapp/api-file-management:$tag .
    # If second param is 1, push the built tag
	if [ ! -z "$2" ] && [ "$2" == 1 ]; then
	    sudo docker push jinnapp/api-file-management:$tag
	fi
	# If second param is 2, push all
	if [ ! -z "$2" ] && [ "$2" == 2 ]; then
	    sudo docker build -t jinnapp/api-file-management .
	    sudo docker push jinnapp/api-file-management:$tag
	    sudo docker push jinnapp/api-file-management:latest
	fi
fi