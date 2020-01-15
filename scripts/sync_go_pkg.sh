#!/bin/bash

# 在 aliyun 的服务器上 go get 经常不通, 所以需要将本地开发环境的 pkg 同步到 aliyun 服务器上

# enable when debug
# set -x

# quit when any error happens
set -e
# quit when use undefined variable
set -u
# quit when error happens in pipe
set -o pipefail
trap "echo 'error: Script failed: see failed command above'" ERR

# make a tar package of local go/pkg/mod
SRC_GO_ROOT=~/go
TARFILE=/tmp/go_pkg_mod.tar.gz
DEST=dev
DEST_GO_ROOT=/root/go

cd $SRC_GO_ROOT/pkg/mod
tar --exclude cache -czf $TARFILE .

scp $TARFILE $DEST:$TARFILE
ssh $DEST rm -rf $DEST_GO_ROOT/pkg/mod/*
ssh $DEST tar -C $DEST_GO_ROOT/pkg/mod -xzf $TARFILE

GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "${GREEN}SUCCESS${NC}"

