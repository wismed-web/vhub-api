#!/bin/bash

set -e

R=`tput setaf 1`
G=`tput setaf 2`
Y=`tput setaf 3`
W=`tput sgr0`

#######################################################################################

# update src swagger comment.
date '+%m-%d-%Y %H:%M:%S' | xargs -I {} sed -i -r 's/Updated@.*/Updated@ {}/' ./server/main.go

# if [[ -z "$1" ]]
# then 
#     echo "${R}need swagger page test access [IP] or [Domain] is needed, abort.${R}"
#     exit
# fi

# $1: swagger test access IP or domain
# if [[ $1 ]]; then
#     sed -i -r "s/@host .+/@host $1/" ./server/main.go
# fi

# update swagger doc
./update.sh

#######################################################################################

cd ./server

# create version info as hard-coded
cd ./auto-gen
go run .
cd -

LDFLAGS="-s -w"
TM=`date +%F@%T@%Z`
OUT=vhub-api\($TM\)

GOARCH=amd64

# For Docker, one build below for linux64 is enough.
OUTPATH_LINUX=./build/linux64/
mkdir -p $OUTPATH_LINUX
CGO_ENABLED=0 GOOS="linux" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $OUT
mv $OUT $OUTPATH_LINUX
echo "${G}server(linux64) built${W}"

#######################################################################################

# copy email config file to current folder
cp -rf ../sendgrid-config.json $OUTPATH_LINUX

# copy init-admin config file to current folder
cp -rf ../init-admin.json $OUTPATH_LINUX

#######################################################################################

# if [[ $1 == 'release' || $1 == 'rel' ]]
# then
#     RELEASE_NAME=wisite-api\($TM\).tar.gz 
#     cd ./build
#     echo $RELEASE_NAME
#     tar -czvf ./$RELEASE_NAME --exclude='./linux64/data' ./linux64
# fi