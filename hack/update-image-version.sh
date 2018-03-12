#!/usr/bin/env sh

FILENAME=$1
IMAGE=$2
VERSION=$3

sed -i -E "s@image: $IMAGE.+@image: $IMAGE:$VERSION@" "$FILENAME"
