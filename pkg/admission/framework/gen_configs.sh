#!/usr/bin/env bash

MAKER_DIR="../../../cmd/cfgmaker"
OUTPUT_DIR="./configs"

rm -rf ${OUTPUT_DIR}

go run ${MAKER_DIR}/maker.go \
	-templatePath ${MAKER_DIR}/object.go.yaml \
	-configPath ${MAKER_DIR}/config.yaml \
	-outputPath ${OUTPUT_DIR}
