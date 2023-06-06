#!/bin/bash
USERNAME="gurkengewuerz"
PROJECT="shoutrrr-api"
REGISTRY="ghcr.io"

docker build --no-cache -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest .
docker push ${REGISTRY}/${USERNAME}/${PROJECT}:latest

echo -e "Done!"
