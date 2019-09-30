#!/usr/bin/env bash

MAKER_DIR="."
OUTPUT_DIR="../../pkg/admission/framework/review/apis"

rm -rf ${OUTPUT_DIR}

go run ${MAKER_DIR}/maker.go \
	-templatePath ${MAKER_DIR}/templates.gohtml \
	-configPath ${MAKER_DIR}/config.yaml \
	-outputPath ${OUTPUT_DIR}

gofmt -w -s -e -d ${OUTPUT_DIR}
