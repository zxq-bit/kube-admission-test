#!/usr/bin/env bash

source ./common-vars.sh

#EXEC_PATH="../../bin/admin"
#
#${EXEC_PATH} \
#    --env-kube-config="${HOME}/.kube/config" \
#    --env-mongo-addr="127.0.0.1:${MONGO_LOCAL_PORT}" \
#    --port=2333

docker run --rm \
    --network="host" \
    -v ${LOCAL_KUBE_CONFIG_PATH}:/root/.kube/host_kube_config \
    -e SERVER_KUBE_CONFIG=/root/.kube/host_kube_config \
    -e SERVER_SERVICE_NAMESPACE=default \
    -e SERVER_SERVICE_NAME=admission-test \
    -e SERVER_SERVICE_SELECTOR=zxq-app:admission-test \
    ${REG_PREFIX}/kube-admission-test-server:${IMAGE_VER}
