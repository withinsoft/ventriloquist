#!/bin/sh

set -e
set -x

docker build -t within/ventriloquist:$(git rev-parse HEAD) . 
docker push within/ventriloquist:$(git rev-parse HEAD)
