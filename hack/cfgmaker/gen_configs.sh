#!/usr/bin/env bash

MAKER_DIR="."
OUTPUT_PATH="../../pkg/admission/framework/review/gen/gen.go"

go run ${MAKER_DIR}/maker.go \
	-templatePath ${MAKER_DIR}/templates.gohtml \
	-configPath ${MAKER_DIR}/config.yaml \
	-outputPath ${OUTPUT_PATH}

gofmt -w -s -e ${OUTPUT_PATH}
