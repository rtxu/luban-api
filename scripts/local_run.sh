
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

cd $ABSPath/../localEnv
go build -o main ../main.go
./main

