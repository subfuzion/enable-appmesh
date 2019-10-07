#!/usr/bin/env bash
# vim:syn=sh:ts=4:sw=4:et:ai

set -ex

if [ -z $COLOR_TELLER_IMAGE ]; then
    echo "COLOR_TELLER_IMAGE environment variable is not set"
    exit 1
fi

docker build -t $COLOR_TELLER_IMAGE .
docker push $COLOR_TELLER_IMAGE
