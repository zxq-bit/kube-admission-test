#!/usr/bin/env bash

LOCAL_KUBE_CONFIG_PATH="${HOME}/.kube/config"

REG_PREFIX="cargo.dev.caicloud.xyz/release"

GIT_VER=`git rev-parse --short HEAD`
GIT_TAG_VER=`git describe --tags --always --dirty`
VERSION=${GIT_TAG_VER}
IMAGE_VER=${GIT_TAG_VER}

touch ./setting.sh
source ./setting.sh
