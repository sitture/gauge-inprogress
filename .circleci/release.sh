#!/usr/bin/env bash

if [ -z "${GITHUB_TOKEN}" ]; then
  echo "GITHUB_TOKEN is not set"
  exit 1
fi

go run build/make.go --all-platforms
go run build/make.go --all-platforms --distro

cd deploy/
for i in `ls`; do
    ${GOPATH}/bin/github-release upload \
        -u ${CIRCLE_PROJECT_USERNAME} \
        -r ${CIRCLE_PROJECT_REPONAME} \
        -t ${CIRCLE_TAG} \
        -n $i -f $i
    if [ $? -ne 0 ];then
        exit 1
    fi
done