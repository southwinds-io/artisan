#!/bin/bash
# removing binaries with root owner
echo removing previous binaries...
rm -rf bin > /dev/null
rm -f *.tar.gz > dev/null
mkdir -p bin/darwin/amd64
mkdir -p bin/darwin/arm64
mkdir -p bin/linux/amd64
# build all binaries
echo building binaries...
art run build-all
# changes the ownership to root before tarring
echo changing binary ownership...
sudo chown root bin/darwin/amd64/art
sudo chown root bin/darwin/arm64/art
sudo chown root bin/linux/amd64/art
# tarring binaries
echo tarring binaries...
tar -zcvf art_linux_amd64.tar.gz -C bin/linux/amd64 .
tar -zcvf art_darwin_amd64.tar.gz -C bin/darwin/amd64 .
tar -zcvf art_darwin_arm64.tar.gz -C bin/darwin/arm64 .
