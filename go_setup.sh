#!/bin/bash

set -x
sudo rm -rf /usr/local/go/
curl -O https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.8.linux-amd64.tar.gz
go get -u github.com/golang/dep/...
go get github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
cd ${HOME}/.go_workspace/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
git checkout $CIRCLE_BRANCH
dep ensure
