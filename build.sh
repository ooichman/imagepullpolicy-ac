#!/bin/bash

if [[ -z $1 ]]; then
	echo "No Image Set asargument 1"
	exit 1
else
	buildah bud -f Dockerfile -t $1 && \
	buildah push $1
fi