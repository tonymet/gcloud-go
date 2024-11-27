#!/bin/zsh
owner=tonymet
repo="gcloud-go"
VERSION=$(date +%Y-%m-%d)-${COMMIT}
build(){
    make build/gcloud-go.tgz
    if [[ $? -ne 0 ]]; then
        echo "ERROR building"
        exit 2
    fi
}