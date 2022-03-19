#!/bin/bash

export TAG="latest"
echo $1
if [ $1 ] ; then
  TAG=$1
fi

echo Building contextcloud/ccb-cli:$TAG

docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t contextcloud/ccb-cli:$TAG .

if [ $? == 0 ] ; then
  docker create --name ccb-cli contextcloud/ccb-cli:$TAG && \
  docker cp ccb-cli:/usr/bin/ccb . && \
  docker rm -f ccb-cli
else
 exit 1
fi