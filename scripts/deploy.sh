#!/bin/bash

# 在 aliyun 的服务器上 go get 经常不通, 所以需要将本地开发环境的 pkg 同步到 aliyun 服务器上

# enable when debug
#set -x

# quit when any error happens
set -e
# quit when use undefined variable
set -u
# quit when error happens in pipe
set -o pipefail
trap "echo 'error: Script failed: see failed command above'" ERR

readonly ABSPath=$(cd $(dirname $0); pwd)
readonly TARFILE=/tmp/luban-api.code.tar.gz
readonly DEST=dev
readonly DEST_GO_ROOT=/root/go
readonly DEST_BIN_ROOT=/luban/app/luban-api
readonly DEST_BIN=${DEST_BIN_ROOT}/main

# 将本地代码打包
pushd $ABSPath/..
tar --exclude scripts -czf $TARFILE .
popd

# copy 一份到部署机
scp $TARFILE $DEST:$TARFILE
ssh $DEST rm -rf $DEST_BIN_ROOT/src/*
ssh $DEST tar -C $DEST_BIN_ROOT/src -xzf $TARFILE

# compile the new binary
ssh $DEST "cd ${DEST_BIN_ROOT}/src && /usr/local/go/bin/go install"

# 停掉老 binary
# error happened when binary has already been stopped, so ignore
set +e
ssh $DEST pkill --full --exact ${DEST_BIN}
set -e

# mv the new binary to the app binary location
ssh $DEST "mv ${DEST_GO_ROOT}/bin/luban-api ${DEST_BIN}"

# 启动新 binary
ssh $DEST "nohup ${DEST_BIN} >& ${DEST_BIN_ROOT}/nohup.log &"

GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "${GREEN}SUCCESS${NC}"

