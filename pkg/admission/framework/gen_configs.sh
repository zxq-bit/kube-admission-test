#!/usr/bin/env bash

MAKER_DIR="../../../hack/cfgmaker"
OUTPUT_DIR="./interfaces/apis"

rm -rf ${OUTPUT_DIR}

go run ${MAKER_DIR}/maker.go \
	-templatePath ${MAKER_DIR}/templates.gohtml \
	-configPath ${MAKER_DIR}/config.yaml \
	-outputPath ${OUTPUT_DIR}

gofmt -w -s -e -d ${OUTPUT_DIR}
