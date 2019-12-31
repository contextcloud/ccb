#!/bin/bash

export TAG="latest"
echo $1
if [ $1 ] ; then
  TAG=$1
fi

echo Building contextgg/faas-cd:$TAG

docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t contextgg/faas-cd:$TAG .

if [ $? == 0 ] ; then
  docker create --name faas-cd contextgg/faas-cd:$TAG && \
  docker cp faas-cd:/usr/bin/faas-cd . && \
  docker rm -f faas-cd
else
 exit 1
fi