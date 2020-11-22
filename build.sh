#!/bin/bash
set -e

make bin 
docker build -t registry.cn-shenzhen.aliyuncs.com/oars/oars-cloud -f Dockerfile .
docker push registry.cn-shenzhen.aliyuncs.com/oars/oars-cloud
