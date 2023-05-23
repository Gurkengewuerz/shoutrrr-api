#!/bin/bash
USERNAME="gurkengewuerz"
PROJECT="shoutrrr-api"
REGISTRY="reg.mc8051.de"

docker build --no-cache -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest .
docker push ${REGISTRY}/${USERNAME}/${PROJECT}:latest

echo -e "Done!"
